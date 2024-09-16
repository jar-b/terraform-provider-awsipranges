// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = ContainsFunction{}
)

func NewContainsFunction() function.Function {
	return ContainsFunction{}
}

type ContainsFunction struct{}

func (r ContainsFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "contains"
}

func (r ContainsFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Contains function",
		MarkdownDescription: "Checks whether an IP address is in an AWS range.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "ip",
				MarkdownDescription: "IP address to search for",
			},
		},
		Return: function.StringReturn{},
	}
}

func (r ContainsFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var data string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &data))
	if resp.Error != nil {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, data))
}
