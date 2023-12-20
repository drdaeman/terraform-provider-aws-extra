package provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceListenerRulesType struct{}

// GetSchema - Returns data source schema
func (r dataSourceListenerRulesType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"listener_arn": {
				Type:        types.StringType,
				Required:    true,
				Description: "ARN of an ELB listener",
			},
			"rules": {
				Type: types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"rule_arn":   types.StringType,
					"is_default": types.BoolType,
					"priority":   types.StringType,
					"conditions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
						"field":                      types.StringType,
						"host_header_config":         types.ListType{ElemType: types.StringType},
						"http_header_config":         types.ListType{ElemType: types.StringType},
						"http_request_method_config": types.ListType{ElemType: types.StringType},
						"path_pattern_config":        types.ListType{ElemType: types.StringType},
						"query_string_config": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
							"key":   types.StringType,
							"value": types.StringType,
						}}},
						"source_ip_config": types.ListType{ElemType: types.StringType},
						"values":           types.ListType{ElemType: types.StringType},
					}}},
					"actions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
						"type":  types.StringType,
						"order": types.NumberType,
						"forward_config": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
							"target_group_arn": types.StringType,
							"weight":           types.NumberType,
						}}},
						"fixed_response_config": types.ObjectType{AttrTypes: map[string]attr.Type{
							"status_code":  types.StringType,
							"message_body": types.StringType,
							"content_type": types.StringType,
						}},
						"target_group_arn": types.StringType,
					}}},
				}}},
				Computed:    true,
				Description: "List of ELB rules",
			},
		},
	}, nil
}

type dataSourceListenerRulesData struct {
	ListenerArn types.String   `tfsdk:"listener_arn"`
	Rules       []listenerRule `tfsdk:"rules"`
}

type listenerRule struct {
	RuleArn    string                   `tfsdk:"rule_arn"`
	IsDefault  bool                     `tfsdk:"is_default"`
	Priority   string                   `tfsdk:"priority"`
	Conditions []listenerRuleConditions `tfsdk:"conditions"`
	Actions    []listenerRuleActions    `tfsdk:"actions"`
}

type listenerRuleConditions struct {
	Field                   string         `tfsdk:"field"`
	HostHeaderConfig        []string       `tfsdk:"host_header_config"`
	HttpHeaderConfig        []string       `tfsdk:"http_header_config"`
	HttpRequestMethodConfig []string       `tfsdk:"http_request_method_config"`
	PathPatternConfig       []string       `tfsdk:"path_pattern_config"`
	QueryStringConfig       []keyValuePair `tfsdk:"query_string_config"`
	SourceIpConfig          []string       `tfsdk:"source_ip_config"`
	Values                  []string       `tfsdk:"values"`
}

type keyValuePair struct {
	Key   string `tfsdk:"key"`
	Value string `tfsdk:"value"`
}

type actionForwardConfig struct {
	TargetGroupArn string `tfsdk:"target_group_arn"`
	Weight         int    `tfsdk:"weight"`
}

type actionFixedResponseConfig struct {
	StatusCode  string `tfsdk:"status_code"`
	MessageBody string `tfsdk:"message_body"`
	ContentType string `tfsdk:"content_type"`
}

type listenerRuleActions struct {
	Type                string                     `tfsdk:"type"`
	Order               int                        `tfsdk:"order"`
	ForwardConfig       []actionForwardConfig      `tfsdk:"forward_config"`
	FixedResponseConfig *actionFixedResponseConfig `tfsdk:"fixed_response_config"`
	TargetGroupArn      *string                    `tfsdk:"target_group_arn"`
}

// NewDataSource creates new data source instance
func (r dataSourceListenerRulesType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceListenerRules{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceListenerRules struct {
	p provider
}

// Read resource information
func (r dataSourceListenerRules) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var cfg dataSourceListenerRulesData
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	elb := elasticloadbalancingv2.New(elasticloadbalancingv2.Options{
		Credentials: r.p.credentials,
		Region:      r.p.region,
		RetryMode:   aws.RetryModeAdaptive,
	})
	res, err := elb.DescribeRules(ctx, &elasticloadbalancingv2.DescribeRulesInput{
		ListenerArn: &cfg.ListenerArn.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to describe ELB rules",
			"DescribeRules call failed:\n\n"+err.Error(),
		)
		return
	}

	for _, rule := range res.Rules {
		conditions := make([]listenerRuleConditions, 0)
		for _, condition := range rule.Conditions {
			var hostHeaderConfig []string
			if condition.HostHeaderConfig != nil {
				hostHeaderConfig = condition.HostHeaderConfig.Values
			}

			var httpHeaderConfig []string
			if condition.HttpHeaderConfig != nil {
				httpHeaderConfig = condition.HttpHeaderConfig.Values
			}

			var httpRequestMethodConfig []string
			if condition.HttpRequestMethodConfig != nil {
				httpRequestMethodConfig = condition.HttpRequestMethodConfig.Values
			}

			var pathPatternConfig []string
			if condition.PathPatternConfig != nil {
				pathPatternConfig = condition.PathPatternConfig.Values
			}

			var queryStringConfig []keyValuePair
			if condition.QueryStringConfig != nil {
				queryStringConfig = make([]keyValuePair, 0)
				for _, elem := range condition.QueryStringConfig.Values {
					queryStringConfig = append(queryStringConfig, keyValuePair{
						Key:   *elem.Key,
						Value: *elem.Value,
					})
				}
			}

			var sourceIpConfig []string
			if condition.SourceIpConfig != nil {
				sourceIpConfig = condition.SourceIpConfig.Values
			}

			conditions = append(conditions, listenerRuleConditions{
				Field:                   *condition.Field,
				HostHeaderConfig:        hostHeaderConfig,
				HttpHeaderConfig:        httpHeaderConfig,
				HttpRequestMethodConfig: httpRequestMethodConfig,
				PathPatternConfig:       pathPatternConfig,
				QueryStringConfig:       queryStringConfig,
				SourceIpConfig:          sourceIpConfig,
				Values:                  condition.Values,
			})
		}

		actions := make([]listenerRuleActions, 0)
		for _, action := range rule.Actions {
			var forwardConfig []actionForwardConfig
			if action.ForwardConfig != nil {
				forwardConfig = make([]actionForwardConfig, 0)
				for _, elem := range action.ForwardConfig.TargetGroups {
					forwardConfig = append(forwardConfig, actionForwardConfig{
						TargetGroupArn: *elem.TargetGroupArn,
						Weight:         int(*elem.Weight),
					})
				}
			}

			var fixedResponseConfig *actionFixedResponseConfig
			if action.FixedResponseConfig != nil {
				fixedResponseConfig = &actionFixedResponseConfig{
					StatusCode:  *action.FixedResponseConfig.StatusCode,
					MessageBody: *action.FixedResponseConfig.MessageBody,
					ContentType: *action.FixedResponseConfig.ContentType,
				}
			}

			actions = append(actions, listenerRuleActions{
				Type:                string(action.Type),
				Order:               int(*action.Order),
				ForwardConfig:       forwardConfig,
				FixedResponseConfig: fixedResponseConfig,
				TargetGroupArn:      action.TargetGroupArn,
			})
		}

		cfg.Rules = append(cfg.Rules, listenerRule{
			RuleArn:    *rule.RuleArn,
			IsDefault:  rule.IsDefault,
			Priority:   *rule.Priority,
			Conditions: conditions,
			Actions:    actions,
		})
	}
	diags = resp.State.Set(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
}
