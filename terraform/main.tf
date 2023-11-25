resource "aws_lambda_function" "line-gpt-rag-function" {
  function_name = "LineGptRagFunction"
  role          = aws_iam_role.lambda_exec.arn
  handler       = "main"
  runtime       = "go1.x"
  filename      = "../main.zip"

  environment {
    variables = {
      OPENAI_API_KEY = var.openai_api_key
    }
  }
}

resource "aws_lambda_function_url" "gpt-rag-function-url" {
  function_name = aws_lambda_function.line-gpt-rag-function.function_name
  authorization_type = "NONE"
}

output "lambda_function_url" {
  value = aws_lambda_function_url.gpt-rag-function-url.function_url
}

resource "aws_iam_role" "lambda_exec" {
  name = "lambda_exec_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        },
      },
    ],
  })
}

# IAM Role for CloudWatch Logs
resource "aws_iam_role" "lambda_logs" {
  name = "lambda_logs_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        },
      },
    ],
  })
}

variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
}