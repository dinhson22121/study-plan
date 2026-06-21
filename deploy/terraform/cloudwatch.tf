###############################################################################
# CloudWatch: app log group + starter alarms (ALB 5xx, EC2 CPU).
# alarm actions fire only when alarm_sns_topic_arn is provided.
###############################################################################

resource "aws_cloudwatch_log_group" "app" {
  name              = "/${var.project}/${var.environment}/app"
  retention_in_days = var.log_retention_days

  tags = { Name = "${local.name}-app-logs" }
}

locals {
  alarm_actions = var.alarm_sns_topic_arn == "" ? [] : [var.alarm_sns_topic_arn]
}

# ---------------------------------------------------------------------------
# ALB 5xx (target-generated) alarm.
# ---------------------------------------------------------------------------

resource "aws_cloudwatch_metric_alarm" "alb_5xx" {
  alarm_name          = "${local.name}-alb-5xx"
  alarm_description   = "ALB target 5xx responses exceed threshold."
  namespace           = "AWS/ApplicationELB"
  metric_name         = "HTTPCode_Target_5XX_Count"
  statistic           = "Sum"
  period              = 300
  evaluation_periods  = 1
  threshold           = var.alb_5xx_threshold
  comparison_operator = "GreaterThanThreshold"
  treat_missing_data  = "notBreaching"

  dimensions = {
    LoadBalancer = aws_lb.main.arn_suffix
    TargetGroup  = aws_lb_target_group.app.arn_suffix
  }

  alarm_actions = local.alarm_actions
  ok_actions    = local.alarm_actions

  tags = { Name = "${local.name}-alb-5xx" }
}

# ---------------------------------------------------------------------------
# EC2 high CPU alarm.
# ---------------------------------------------------------------------------

resource "aws_cloudwatch_metric_alarm" "ec2_cpu" {
  alarm_name          = "${local.name}-ec2-cpu-high"
  alarm_description   = "EC2 app host average CPU is high."
  namespace           = "AWS/EC2"
  metric_name         = "CPUUtilization"
  statistic           = "Average"
  period              = 300
  evaluation_periods  = 2
  threshold           = var.ec2_cpu_threshold_percent
  comparison_operator = "GreaterThanThreshold"
  treat_missing_data  = "notBreaching"

  dimensions = {
    InstanceId = aws_instance.app.id
  }

  alarm_actions = local.alarm_actions
  ok_actions    = local.alarm_actions

  tags = { Name = "${local.name}-ec2-cpu-high" }
}

# ---------------------------------------------------------------------------
# No healthy ALB targets alarm (catches a crashed app host).
# ---------------------------------------------------------------------------

resource "aws_cloudwatch_metric_alarm" "alb_unhealthy_hosts" {
  alarm_name          = "${local.name}-alb-unhealthy-hosts"
  alarm_description   = "ALB has zero healthy targets."
  namespace           = "AWS/ApplicationELB"
  metric_name         = "HealthyHostCount"
  statistic           = "Minimum"
  period              = 60
  evaluation_periods  = 3
  threshold           = 1
  comparison_operator = "LessThanThreshold"
  treat_missing_data  = "breaching"

  dimensions = {
    LoadBalancer = aws_lb.main.arn_suffix
    TargetGroup  = aws_lb_target_group.app.arn_suffix
  }

  alarm_actions = local.alarm_actions
  ok_actions    = local.alarm_actions

  tags = { Name = "${local.name}-alb-unhealthy-hosts" }
}
