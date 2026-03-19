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
)

var (
	_ datasource.DataSource = &labelBuilderBaseDataSource{}
)

func NewLabelMetadataDataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource {
			return newLabelMetadataDataSource(
				ctxmodel.ContextTypeNamespace,
				"namespace",
			)
		},
	}
}

func newLabelMetadataDataSource(labelId ctxmodel.ContextType, typeName string) datasource.DataSource {
	return &labelBuilderBaseDataSource{
		LabelId:  labelId,
		TypeName: typeName,
	}
}

type labelBuilderBaseDataSource struct {
	LabelId  ctxmodel.ContextType
	TypeName string
}

type labelBuilderDataSourceBaseSchema struct {
	Name                  types.String                             `tfsdk:"name"`
	Context               ctxschema.ContextSchema                  `tfsdk:"context"`
	Vars                  types.Map                                `tfsdk:"vars"`
	Mappers               *[]ctxschema.ContextMapperFunctionSchema `tfsdk:"mappers"`
	IdCasing              types.String                             `tfsdk:"id_casing"`
	IdPrefix              types.String                             `tfsdk:"id_prefix"`
	IncludeResourceTypeInId types.Bool                             `tfsdk:"include_resource_type_in_id"`
}

func (d *labelBuilderBaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("context_%s", d.TypeName)
}

func (d *labelBuilderBaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"context": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"stack": schema.ListNestedAttribute{
						Required: true,
						Validators: []validator.List{
							ctxvalidator.ContextStackOrderValidator(d.LabelId),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: contextStackElementAttributes(),
						},
					},
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
			"id_casing": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					ctxvalidator.IdCasingValidator(),
				},
			},
			"id_prefix": schema.StringAttribute{
				Optional: true,
			},
			"include_resource_type_in_id": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (d *labelBuilderBaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var dataSource labelBuilderDataSourceBaseSchema

	diags := req.Config.Get(ctx, &dataSource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataSource.Context.Stack.Add(
		dataSource.Name,
		d.LabelId,
		dataSource.Vars,
		dataSource.Mappers,
		dataSource.IdCasing,
		dataSource.IdPrefix,
		dataSource.IncludeResourceTypeInId,
	)

	diags = resp.State.Set(ctx, &dataSource)
	resp.Diagnostics.Append(diags...)
}
