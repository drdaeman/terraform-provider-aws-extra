package main

import (
	"context"
	"github.com/drdaeman/terraform-provider-aws-extras/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func main() {
	tfsdk.Serve(context.Background(), provider.New, tfsdk.ServeOpts{
		Name: "aws-extras",
	})
}
