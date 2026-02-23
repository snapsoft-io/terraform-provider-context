// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxschema"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxvalidator"
)

var _ provider.Provider = &contextProvider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &contextProvider{
			version: version,
		}
	}
}

type contextProvider struct {
	version string
}

type contextProviderSchema struct {
	Mappers         *[]ctxschema.ContextMapperFunctionSchema `tfsdk:"mappers"`
	MappersFilePath types.String                             `tfsdk:"mappers_file_path"`
	Vars            types.Map                                `tfsdk:"vars"`
}

func (p *contextProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "context"
	resp.Version = p.version
}

func (p *contextProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"mappers": schema.ListNestedAttribute{
				Optional: true,
				Validators: []validator.List{
					ctxvalidator.JqSyntaxValidator(),
				},
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
			"mappers_file_path": schema.StringAttribute{
				Optional: true,
			},
			"vars": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (p *contextProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config contextProviderSchema

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mappers = make([]ctxmodel.ContextMapperFunctionModel, 0)
	if config.Mappers != nil {
		for _, mapper := range *config.Mappers {
			mappers = append(mappers, *mapper.ToModel())
		}
	} else if !config.MappersFilePath.IsNull() {
		mappersJson, err := ctxmodel.NewContextMapperFunctionModelListFromJson(config.MappersFilePath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Failed to read the '%s' mappers JSON file", config.MappersFilePath.ValueString()), err.Error())
			return
		}

		mappers = *mappersJson
	}

	var vars = make(map[string]string)
	if !config.Vars.IsNull() {
		diag := config.Vars.ElementsAs(ctx, &vars, false)
		if diag.HasError() {
			resp.Diagnostics.AddError(fmt.Sprintf("conversion error: %v", diag.Errors()), "")
			return
		}
	}

	resp.DataSourceData = &ctxmodel.ContextProviderConfigModel{
		MapperFunctions: &mappers,
		Vars:            vars,
	}
}

func (p *contextProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *contextProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return append(
		NewLabelMetadataDataSources(),
		NewItemDataSource,
		NewVariableDataSource)
}
