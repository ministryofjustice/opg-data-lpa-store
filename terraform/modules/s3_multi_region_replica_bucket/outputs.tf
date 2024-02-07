output "bucket" {
  value = aws_s3_bucket.bucket
}

output "encryption_kms_key" {
  value = aws_kms_key.s3
}
