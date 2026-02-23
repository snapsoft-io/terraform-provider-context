// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxevaluator"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxschema"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxvalidator"
	"github.com/snapsoft/terraform-provider-context/internal/utils"
)

var (
	_ datasource.DataSource = &itemDataSource{}
)

func NewItemDataSource() datasource.DataSource {
	return &itemDataSource{}
}

type itemDataSource struct {
	providerConfig *ctxmodel.ContextProviderConfigModel
}

type itemDataSourceModel struct {
	Name         types.String            `tfsdk:"name"`
	ResourceType types.String            `tfsdk:"resource_type"`
	Context      ctxschema.ContextSchema `tfsdk:"context"`
	Id           types.String            `tfsdk:"id"`
	Tags         types.Map               `tfsdk:"tags"`
}

func (d *itemDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "context_label_item"
}

func (d *itemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"resource_type": schema.StringAttribute{
				Required: true,
			},
			"context": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"stack": schema.ListNestedAttribute{
						Required: true,
						Validators: []validator.List{
							ctxvalidator.ContextStackOrderValidator(ctxmodel.ContextLabelItem),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required: true,
								},
								"label_id": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										ctxvalidator.ContextLabelIdValueValidator(),
									},
								},
								"vars": schema.MapAttribute{
									Optional:    true,
									ElementType: types.StringType,
								},
								"mappers": schema.ListNestedAttribute{
									Optional: true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Required: true,
											},
											"run_condition": schema.StringAttribute{
												Optional: true,
											},
											"function": schema.StringAttribute{
												Required: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"tags": schema.MapAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *itemDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*ctxmodel.ContextProviderConfigModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ContextProviderConfigModel, but got: %T.", req.ProviderData),
		)

		return
	}

	d.providerConfig = providerConfig
}

func (d *itemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var dataSource itemDataSourceModel

	diags := req.Config.Get(ctx, &dataSource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add resource_type as variable to the item stack element
	itemVars, diags := types.MapValue(types.StringType, map[string]attr.Value{"resource_type": dataSource.ResourceType})
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	dataSource.Context.Stack.AddWithNameLabelVars(dataSource.Name, ctxmodel.ContextLabelItem, itemVars)

	// Collect context data and evaluate the mappers on it
	stack := dataSource.Context.Stack.ToAnyGoType(ctx)
	vars := dataSource.Context.Stack.GetStackVarsInTopDownOrder(ctx)
	vars = utils.MergeMaps(d.providerConfig.Vars, vars)
	extraJqParams := map[string]any{
		"stack": stack,
		"vars":  utils.ToAnyMap(vars),
	}
	mappers := dataSource.Context.Stack.GetStackMappersInBottomUpOrder()
	*d.providerConfig.MapperFunctions = append(*mappers, *d.providerConfig.MapperFunctions...)
	evaluatedContextMain, err := ctxevaluator.EvaluateJqMappers(*d.providerConfig.MapperFunctions, extraJqParams)
	if err != nil {
		diags.AddError("Failed to evaluate the context with the given mappers", err.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	// Extract id and tags from evaluated context outputs section
	contextOutput, err := ctxevaluator.EvaluateContextOutput(evaluatedContextMain)
	if err != nil {
		diags.AddError("Failed to retrieve required output values from the 'context.main' data structure", err.Error())
		resp.Diagnostics.Append(diags...)
		return
	}
	tagMap, err := utils.ConvertGoMapToTfMap(contextOutput.Tags)
	if err != nil {
		diags.AddError("Failed to convert go map to terraform map value", err.Error())
		resp.Diagnostics.Append(diags...)
		return
	}
	dataSource.Id = types.StringValue((*contextOutput).Id)
	dataSource.Tags, err = utils.MergeTfMaps(dataSource.Tags, tagMap)
	if err != nil {
		diags.AddError("Failed to merge context tags with currently given tags", err.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	// Return the full context if TF_LOG value is DEBUG
	tfLogEnv := strings.ToUpper(os.Getenv("TF_LOG"))
	if tfLogEnv != "DEBUG" {
		dataSource.Context = *ctxschema.NewEmptyContextModel()
	}

	diags = resp.State.Set(ctx, &dataSource)
	resp.Diagnostics.Append(diags...)
}
