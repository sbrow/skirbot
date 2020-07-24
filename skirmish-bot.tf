provider "heroku" {
  version = "~> 2.0"
}

variable "token" {
  description = "The Discord bot token"
  type        = string
}

resource "heroku_app" "app" {
  name   = "skirmish-bot"
  region = "us"
  buildpacks = ["heroku/go"]

  sensitive_config_vars = {
    TOKEN        = var.token
  }
}

# Create a database, and configure the app to use it
resource "heroku_addon" "database" {
  app  = heroku_app.app.name
  plan = "heroku-postgresql:hobby-dev"
}

/*
resource "heroku_build" "example" {
  app = heroku_app.example.name

  source = {
    url = "https://github.com/sbrow/skirbot/archive/master.tar.gz"
  }
}
*/

resource "heroku_formation" "example" {
  app        = heroku_app.app.name
  type       = "worker"
  quantity   = 1
  size       = "Free"
  #depends_on = [heroku_build.example]
}
