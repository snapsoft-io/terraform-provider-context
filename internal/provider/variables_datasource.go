// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxschema"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxvalidator"
	"github.com/snapsoft/terraform-provider-context/internal/utils"
)

var (
	_ datasource.DataSource = &itemDataSource{}
)

func NewVariableDataSource() datasource.DataSource {
	return &variableDataSource{}
}

type variableDataSource struct {
	providerConfig *ctxmodel.ContextProviderConfigModel
}

type variableDataSourceModel struct {
	Context ctxschema.ContextSchema `tfsdk:"context"`
	Vars    types.Map               `tfsdk:"vars"`
}

func (d *variableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "context_variables"
}

func (d *variableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"vars": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *variableDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *variableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var dataSource variableDataSourceModel

	diags := req.Config.Get(ctx, &dataSource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vars := dataSource.Context.Stack.GetStackVarsInTopDownOrder(ctx)
	vars = utils.MergeMaps(d.providerConfig.Vars, vars)

	tfMap, diags := types.MapValueFrom(ctx, types.StringType, vars)
	if diags.HasError() {
		return
	}
	dataSource.Vars = tfMap

	diags = resp.State.Set(ctx, &dataSource)
	resp.Diagnostics.Append(diags...)
}
