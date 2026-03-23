// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxvalidator"
)

// contextStackElementAttributes returns the shared schema attributes for context stack elements.
// Used consistently across all data sources that accept a context stack.
func contextStackElementAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
		"label_id": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				ctxvalidator.ContextTypeValueValidator(),
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
	}
}
