package provider

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/drdaeman/terraform-provider-aws-extras/validators"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured  bool
	region      string
	credentials aws.CredentialsProvider
}

type providerData struct {
	Region        types.String `tfsdk:"region"`
	AssumeRoleArn types.String `tfsdk:"assume_role_arn"`
	SessionName   types.String `tfsdk:"session_name"`
}

func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"region": {
				Type:        types.StringType,
				Required:    true,
				Description: "AWS Region",
			},
			"assume_role_arn": {
				Type:        types.StringType,
				Optional:    true,
				Description: "ARN of an IAM Role to assume",
			},
			"session_name": {
				Type:        types.StringType,
				Optional:    true,
				Description: "Name for the assumed role session",
				Validators: []tfsdk.AttributeValidator{
					validators.StringLenBetween(2, 64),
					validators.StringMatch(
						regexp.MustCompile(`[\w+=,.@\-]*`),
						"Name must be a string of characters consisting of upper- and lower-case alphanumeric"+
							" characters with no spaces. You can also include underscores or any of"+
							" the following characters: =,.@-",
					),
				},
			},
		},
	}, nil
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var cfg providerData
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to load AWS config",
			"LoadDefaultConfig failed:\n\n"+err.Error(),
		)
		return
	}

	awsCredentials := awsConfig.Credentials
	if !cfg.AssumeRoleArn.Unknown {
		stsClient := sts.NewFromConfig(awsConfig)
		assumeRoleResponse, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
			RoleArn:         &cfg.AssumeRoleArn.Value,
			RoleSessionName: &cfg.SessionName.Value,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to assume role",
				"AssumeRole failed:\n\n"+err.Error(),
			)
			return
		}
		awsCredentials = credentials.NewStaticCredentialsProvider(
			*assumeRoleResponse.Credentials.AccessKeyId,
			*assumeRoleResponse.Credentials.SecretAccessKey,
			*assumeRoleResponse.Credentials.SessionToken,
		)
	}

	p.region = cfg.Region.Value
	p.credentials = awsCredentials
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"awsx_lb_listener_rules": dataSourceListenerRulesType{},
	}, nil
}
