output "bucket" {
  value = aws_s3_bucket.bucket
}

output "kms_key" {
  value = aws_kms_key.key
}
