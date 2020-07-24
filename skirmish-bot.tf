variable "heroku_api_key" {
  description = "API token for authenticating heroku"
  type        = string
}

variable "token" {
  description = "The Discord bot token"
  type        = string
}

variable "code-version" {
  description = "The version of code to run"
  type        = string
}

terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "electricpandafishgames"

    workspaces {
      name = "skirmish-bot"
    }
  }
}

provider "heroku" {
  version = "~> 2.0"
  api_key = var.heroku_api_key
}


resource "heroku_app" "app" {
  name       = "skirmish-bot"
  region     = "us"
  buildpacks = ["heroku/go"]

  sensitive_config_vars = {
    TOKEN = var.token
  }
}

# Create a database, and configure the app to use it
resource "heroku_addon" "database" {
  app  = heroku_app.app.name
  plan = "heroku-postgresql:hobby-dev"
}

resource "heroku_build" "example" {
  app = heroku_app.app.name

  source = {
    # Deploy local code
    # path = "."
    url     = "https://github.com/username/example/archive/${var.code_version}.tar.gz"
    version = var.code_version
  }
}

resource "heroku_formation" "example" {
  app        = heroku_app.app.name
  type       = "worker"
  quantity   = 1
  size       = "Free"
  depends_on = [heroku_build.example]
}
