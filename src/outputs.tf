output "zones" {
  value       = aws_route53_zone.root
  description = "DNS zones"
}

output "acms" {
  value       = { for k, v in module.acm : k => v.arn }
  description = "ACM certificates for domains"
}

output "query_logging_configs" {
  value = {
    for domain, query_log in aws_route53_query_log.default : domain => {
      id            = query_log.id
      arn           = query_log.arn
      log_group_arn = query_log.cloudwatch_log_group_arn
    }
  }
  description = "Route 53 query logging configurations, keyed by domain name"
}
