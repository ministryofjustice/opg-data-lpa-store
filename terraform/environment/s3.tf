resource "aws_iam_role" "s3_replication_role" {
  name               = "s3-replication-role-${local.environment_name}"
  description        = "IAM Role for S3 replication in ${local.environment_name}"
  assume_role_policy = data.aws_iam_policy_document.s3_replication_role_assume_role.json
  provider           = aws.global
}

data "aws_iam_policy_document" "s3_replication_role_assume_role" {
  provider = aws.global
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

module "s3_lpa_store_static_eu_west_1" {
  source                   = "../modules/s3_multi_region_replica_bucket"
  accounts_allowed_to_read = [local.backup_account_id]
  bucket_name              = "opg-lpa-store-static-${local.environment_name}-eu-west-1"
  force_destroy            = local.is_ephemeral
  kms_allowed_iam_roles    = module.eu_west_1.lambda_iam_roles[*].arn
  replication_configuration = concat(
    [{
      account_id  = data.aws_caller_identity.current.account_id,
      bucket      = module.s3_lpa_store_static_eu_west_2.bucket
      kms_key_arn = module.s3_lpa_store_static_eu_west_2.encryption_kms_key.arn
    }],
  local.cross_account_s3_replica_config)
  replication_kms_key_arns = [
    module.s3_lpa_store_static_eu_west_2.encryption_kms_key.arn
  ]
  s3_access_logging_bucket = "s3-access-logs-opg-lpa-store-${local.environment.account_name}-eu-west-1"
  s3_replication_role      = aws_iam_role.s3_replication_role
  providers = {
    aws = aws.eu_west_1
  }
}

module "s3_lpa_store_static_eu_west_2" {
  source                   = "../modules/s3_multi_region_replica_bucket"
  accounts_allowed_to_read = [local.backup_account_id]
  bucket_name              = "opg-lpa-store-static-${local.environment_name}-eu-west-2"
  force_destroy            = local.is_ephemeral
  kms_allowed_iam_roles    = module.eu_west_2.lambda_iam_roles[*].arn
  replication_configuration = concat(
    [{
      account_id  = data.aws_caller_identity.current.account_id,
      bucket      = module.s3_lpa_store_static_eu_west_1.bucket
      kms_key_arn = module.s3_lpa_store_static_eu_west_1.encryption_kms_key.arn
    }],
  local.cross_account_s3_replica_config)
  replication_kms_key_arns = [
    module.s3_lpa_store_static_eu_west_1.encryption_kms_key.arn
  ]
  s3_access_logging_bucket = "s3-access-logs-opg-lpa-store-${local.environment.account_name}-eu-west-2"
  s3_replication_role      = aws_iam_role.s3_replication_role
  providers = {
    aws = aws.eu_west_2
  }
}

module "s3_data_store_backup_account" {
  count                           = local.cross_account_backup_enabled ? 1 : 0
  source                          = "../modules/s3_cross_account_backup"
  bucket_name                     = "opg-lpa-store-static-${local.environment_name}-backup-eu-west-2"
  environment_name                = local.environment_name
  force_destroy                   = local.is_ephemeral
  s3_access_logging_bucket_prefix = "s3-access-logs-opg-opg-backups-opg-backups"
  s3_replication_role             = aws_iam_role.s3_replication_role
  providers = {
    aws.backup-account = aws.opg_backup
    aws.source-account = aws.global
  }
}

locals {
  cross_account_s3_replica_config = local.cross_account_backup_enabled ? [
    {
      account_id  = local.backup_account_id,
      bucket      = module.s3_data_store_backup_account[0].bucket
      kms_key_arn = module.s3_data_store_backup_account[0].kms_key.arn
    }
  ] : []
}
