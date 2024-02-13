variable "name_suffix" {
  description = "Suffix to append to resource names to ensure enviornment/target uniqueness"
  type        = string
}

variable "rule" {
  description = "The rule that triggers this target"
  type = object({
    event_bus_name = string
    name           = string
  })
}

variable "target_event_bus_arn" {
  description = "The ARN of the target event bus"
  type        = string
}

variable "dead_letter_queue_arn" {
  description = "The dead letter queue to send failed messages to"
  type        = string
}
