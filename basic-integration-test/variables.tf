# =============================================================================
# Cloud Foundry Connection Variables
# =============================================================================

variable "cf_api_url" {
  type        = string
  description = "Cloud Foundry API URL"
}

variable "cf_user" {
  type        = string
  description = "Cloud Foundry username"
}

variable "cf_password" {
  type        = string
  sensitive   = true
  description = "Cloud Foundry password"
}

variable "cf_skip_ssl_validation" {
  type        = bool
  default     = true
  description = "Skip SSL certificate validation"
}

# =============================================================================
# Test Configuration Variables
# =============================================================================

variable "org_name" {
  type        = string
  default     = "test-org"
  description = "Name of the organization to create"
}

variable "space_name" {
  type        = string
  default     = "test-space"
  description = "Name of the space to create"
}

variable "app_name" {
  type        = string
  default     = "test-go-app"
  description = "Name of the test application"
}

variable "app_memory" {
  type        = string
  default     = "64M"
  description = "Memory allocation for the test application"
}

variable "app_instances" {
  type        = number
  default     = 1
  description = "Number of application instances"
}

# =============================================================================
# Service Broker Configuration Variables
# =============================================================================

variable "broker_org_name" {
  type        = string
  default     = "service-broker-org"
  description = "Name of the organization for service broker"
}

variable "broker_space_name" {
  type        = string
  default     = "service-broker-space"
  description = "Name of the space for service broker"
}

variable "broker_app_name" {
  type        = string
  default     = "simple-service-broker"
  description = "Name of the service broker application"
}

variable "broker_app_memory" {
  type        = string
  default     = "64M"
  description = "Memory allocation for the service broker application"
}

variable "broker_username" {
  type        = string
  default     = "admin"
  description = "Username for service broker authentication"
}

variable "broker_password" {
  type        = string
  sensitive   = true
  default     = "password"
  description = "Password for service broker authentication"
}

variable "service_broker_name" {
  type        = string
  default     = "dummy-service-broker"
  description = "Name of the registered service broker"
}

variable "service_instance_name" {
  type        = string
  default     = "dummy-service-instance"
  description = "Name of the service instance"
}

variable "service_binding_name" {
  type        = string
  default     = "dummy-service-key"
  description = "Name of the service key binding"
}

# =============================================================================
# Test Users for Role Assignment
# =============================================================================

variable "test_users" {
  type = list(object({
    username = string
    email    = string
    origin   = optional(string, "uaa")
  }))
  default = [
    {
      username = "test-developer@example.com"
      email    = "test-developer@example.com"
    },
    {
      username = "test-manager@example.com"
      email    = "test-manager@example.com"
    }
  ]
  description = "List of test users to create and assign roles"
}
