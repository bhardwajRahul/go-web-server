// Atlas configuration file
env "dev" {
  src = "file://internal/store/schema.sql"
  url = env("DATABASE_URL")
  dev = "docker://postgres/15/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "prod" {
  src = "file://internal/store/schema.sql"
  url = env("DATABASE_URL")
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}