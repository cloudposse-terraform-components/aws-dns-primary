variable "region" {
  type        = string
  description = "AWS Region"
}

variable "domain_names" {
  type        = list(string)
  default     = null
  description = "Root domain name list, e.g. `[\"example.net\"]`"
}

variable "record_config" {
  description = "DNS Record config"
  type = list(object({
    root_zone = string
    name      = string
    type      = string
    ttl       = string
    records   = list(string)
  }))
  default = []
}

# Elastic Load Balancing Hosted Zone IDs can be found here: https://docs.aws.amazon.com/general/latest/gr/elb.html
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_record#alias-record
variable "alias_record_config" {
  description = "DNS Alias Record config"
  type = list(object({
    root_zone              = string
    name                   = string
    type                   = string
    zone_id                = string
    record                 = string
    evaluate_target_health = bool
  }))
  default = []
}

variable "dns_soa_config" {
  type        = string
  description = <<-EOT
    Root domain name DNS SOA record:
    - awsdns-hostmaster.amazon.com. ; AWS default value for administrator email address
    - 1 ; serial number, not used by AWS
    - 7200 ; refresh time in seconds for secondary DNS servers to refresh SOA record
    - 900 ; retry time in seconds for secondary DNS servers to retry failed SOA record update
    - 1209600 ; expire time in seconds (1209600 is 2 weeks) for secondary DNS servers to remove SOA record if they cannot refresh it
    - 60 ; nxdomain TTL, or time in seconds for secondary DNS servers to cache negative responses
    See [SOA Record Documentation](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/SOA-NSrecords.html) for more information.
  EOT
  default     = "awsdns-hostmaster.amazon.com. 1 7200 900 1209600 60"
}

variable "query_logging_enabled" {
  type        = bool
  description = <<-EOT
    Whether to enable Route 53 DNS query logging for every public hosted zone created by this component.
    Creates one CloudWatch Logs log group per zone (named `/aws/route53/<zone name>`) in `query_logging_region`
    and a query logging configuration that publishes to it.
    See [Logging DNS queries](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/query-logs.html).
  EOT
  default     = false
  nullable    = false
}

variable "query_logging_region" {
  type        = string
  description = <<-EOT
    Region in which to create the query logging CloudWatch Logs log groups and resource policy.
    Route 53 only publishes query logs to log groups in `us-east-1`, so change this only for partitions
    with a different global-service Region.
  EOT
  default     = "us-east-1"
  nullable    = false
}

variable "query_logging_retention_in_days" {
  type        = number
  description = "Number of days to retain DNS query logs in CloudWatch Logs. Set to `0` to never expire log events."
  default     = 365
  nullable    = false

  validation {
    condition     = contains([0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653], var.query_logging_retention_in_days)
    error_message = "The query_logging_retention_in_days must be one of the values supported by CloudWatch Logs: 0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653."
  }
}

variable "query_logging_kms_key_arn" {
  type        = string
  description = <<-EOT
    ARN of a KMS key in `query_logging_region` used to encrypt the query logging log groups.
    The key policy must allow the `logs.<region>.amazonaws.com` service principal to use the key.
    If `null`, the log groups use CloudWatch Logs default encryption.
  EOT
  default     = null
}

variable "query_logging_resource_policy_enabled" {
  type        = bool
  description = <<-EOT
    Whether to create the CloudWatch Logs resource policy that allows Route 53 to write to the query logging log groups.
    CloudWatch Logs allows at most 10 account-scoped resource policies per Region; set to `false` if the account already
    has a policy granting `route53.amazonaws.com` write access to `/aws/route53/*` log groups.
  EOT
  default     = true
  nullable    = false
}

variable "query_logging_resource_policy_name" {
  type        = string
  description = "Name of the CloudWatch Logs resource policy for query logging. Defaults to `<module.this.id>-route53-query-logging`."
  default     = null
}
