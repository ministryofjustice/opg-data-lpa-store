lpaDeedService = softwareSystem "LPA Deed Service" "Stores deeds and manages changes to them" {
    database = container "Database" "Stores deeds and events" "DynamoDB" "Database"
    lambda = container "Lambda" "Executes code for storing, updating and retrieving deeds" "AWS Lambda, Go" "Component" {
        -> database "Queries and writes to"
    }
    iam = container "IAM" "Manages permissions to API Gateway" "AWS IAM" "Component"
    certificateManager = container "Certificate Manager" "Generate a valid cert for SSL connectivity to the API" "AWS Certificate Manager" "Component"
    dns = container "DNS" "Generate a friendly DNS Name for the API" "AWS Route 53" "Component"
    apiGateway = container "API Gateway" "Provides a REST API for communication to the service." "AWS API Gateway v2, OpenAPI" "Component" {
        -> lambda "Forwards requests to and Returns responses from"
        -> iam "Validates requests"
    }
}
