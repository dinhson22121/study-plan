###############################################################################
# Route53 + ACM (DNS-validated) for the API domain, attached to the ALB.
###############################################################################

locals {
  api_fqdn = "${var.api_subdomain}.${var.domain_name}"
}

# Either create the hosted zone or look one up by name.
resource "aws_route53_zone" "main" {
  count = var.create_route53_zone ? 1 : 0
  name  = var.domain_name

  tags = { Name = "${local.name}-zone" }
}

data "aws_route53_zone" "main" {
  count        = var.create_route53_zone ? 0 : 1
  name         = var.domain_name
  private_zone = false
}

locals {
  zone_id = var.create_route53_zone ? aws_route53_zone.main[0].zone_id : data.aws_route53_zone.main[0].zone_id
}

# ---------------------------------------------------------------------------
# ACM certificate (DNS validation).
# ---------------------------------------------------------------------------

resource "aws_acm_certificate" "api" {
  domain_name       = local.api_fqdn
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }

  tags = { Name = "${local.name}-cert" }
}

resource "aws_route53_record" "cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.api.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      type   = dvo.resource_record_type
      record = dvo.resource_record_value
    }
  }

  zone_id         = local.zone_id
  name            = each.value.name
  type            = each.value.type
  records         = [each.value.record]
  ttl             = 60
  allow_overwrite = true
}

resource "aws_acm_certificate_validation" "api" {
  certificate_arn         = aws_acm_certificate.api.arn
  validation_record_fqdns = [for r in aws_route53_record.cert_validation : r.fqdn]
}

# ---------------------------------------------------------------------------
# A/ALIAS record: api.<domain> -> ALB.
# ---------------------------------------------------------------------------

resource "aws_route53_record" "api" {
  zone_id = local.zone_id
  name    = local.api_fqdn
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}
