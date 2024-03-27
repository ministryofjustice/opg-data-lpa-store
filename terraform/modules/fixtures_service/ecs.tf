# loadbalancer + dns (loadbalancer in public subnet, application in application subnet)

resource "aws_ecs_cluster" "main" {
  name = "fixtures-${var.environment_name}"

  provider = aws.region
}

resource "aws_ecs_service" "fixtures" {
  name                  = "fixtures"
  cluster               = aws_ecs_cluster.main.id
  task_definition       = aws_ecs_task_definition.fixtures.arn
  desired_count         = 1
  platform_version      = "1.4.0"
  wait_for_steady_state = true
  propagate_tags        = "SERVICE"
  launch_type           = "FARGATE"

  load_balancer {
    target_group_arn = aws_lb_target_group.fixtures.arn
    container_name   = "fixtures"
    container_port   = 80
  }

  network_configuration {
    security_groups  = [aws_security_group.ecs.id]
    subnets          = var.application_subnet_ids
    assign_public_ip = false
  }

  timeouts {
    create = "5m"
    update = "5m"
  }

  provider = aws.region
}

resource "aws_ecs_task_definition" "fixtures" {
  family                   = "fixtures-${var.environment_name}"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  container_definitions    = "[${local.container_definition}]"
  task_role_arn            = aws_iam_role.task_role.arn
  execution_role_arn       = aws_iam_role.execution_role.arn

  provider = aws.region
}

locals {
  container_definition = jsonencode(
    {
      cpu                    = 1,
      essential              = true,
      image                  = var.ecr_image_uri,
      mountPoints            = [],
      readonlyRootFilesystem = true
      name                   = "fixtures",
      portMappings = [
        {
          containerPort = 80,
          hostPort      = 80,
          protocol      = "tcp"
        }
      ],
      volumesFrom = [],
      logConfiguration = {
        logDriver = "awslogs",
        options = {
          awslogs-group         = aws_cloudwatch_log_group.fixtures.name,
          awslogs-region        = data.aws_region.current.name,
          awslogs-stream-prefix = var.environment_name
        }
      },
      secrets = [
        {
          name      = "JWT_SECRET_KEY",
          valueFrom = data.aws_secretsmanager_secret.jwt_secret_key.arn
        }
      ],
      environment = [
        {
          name  = "BASE_URL",
          value = "https://${var.service_url}",
        }
      ]
    }
  )
}
