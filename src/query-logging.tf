# Route 53 DNS query logging for the public hosted zones created by this component.
#
# Route 53 only publishes query logs to CloudWatch Logs log groups in us-east-1,
# and the CloudWatch Logs resource policy that lets Route 53 write to them must
# also live in us-east-1. The resource-level `region` argument (AWS provider v6)
# places these resources there regardless of `var.region`, so no aliased
# provider is needed.
#
# https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/query-logs.html
# https://docs.aws.amazon.com/Route53/latest/APIReference/API_CreateQueryLoggingConfig.html

locals {
  query_logging_enabled         = module.this.enabled && var.query_logging_enabled
  query_logging_domains         = local.query_logging_enabled ? local.domains_set : []
  query_logging_policy_enabled  = local.query_logging_enabled && var.query_logging_resource_policy_enabled
  query_logging_policy_name     = coalesce(var.query_logging_resource_policy_name, join("-", compact([module.this.id, "route53-query-logging"])))
  query_logging_source_accounts = local.query_logging_enabled ? [one(data.aws_caller_identity.current[*].account_id)] : []
}

data "aws_caller_identity" "current" {
  count = local.query_logging_enabled ? 1 : 0
}

resource "aws_cloudwatch_log_group" "query_logging" {
  for_each = local.query_logging_domains

  region = var.query_logging_region

  name              = "/aws/route53/${each.value}"
  retention_in_days = var.query_logging_retention_in_days
  kms_key_id        = var.query_logging_kms_key_arn
  tags              = module.this.tags
}

data "aws_iam_policy_document" "query_logging" {
  count = local.query_logging_policy_enabled ? 1 : 0

  statement {
    sid    = "Route53QueryLogging"
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [for log_group in aws_cloudwatch_log_group.query_logging : "${log_group.arn}:*"]

    principals {
      type        = "Service"
      identifiers = ["route53.amazonaws.com"]
    }

    # Confused-deputy protection: only this account's query-logged zones may write.
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = local.query_logging_source_accounts
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = [for domain in local.query_logging_domains : aws_route53_zone.root[domain].arn]
    }
  }
}

# Account-scoped policy: CloudWatch Logs allows at most 10 of these per Region.
# Set `query_logging_resource_policy_enabled = false` to manage the policy elsewhere.
resource "aws_cloudwatch_log_resource_policy" "query_logging" {
  count = local.query_logging_policy_enabled ? 1 : 0

  region = var.query_logging_region

  policy_name     = local.query_logging_policy_name
  policy_document = one(data.aws_iam_policy_document.query_logging[*].json)
}

resource "aws_route53_query_log" "default" {
  for_each = aws_cloudwatch_log_group.query_logging

  zone_id                  = aws_route53_zone.root[each.key].zone_id
  cloudwatch_log_group_arn = each.value.arn

  depends_on = [aws_cloudwatch_log_resource_policy.query_logging]
}
