workspace {

  model {
    user = person "Developer" {
      description "Uses the system to manage database migrations."
    }

    cicd = softwareSystem "CI/CD System" {
      description "Triggers automatic migrations during deployments."
    }

    system = softwareSystem "Database Migration Service" {
      description "Compares schemas, generates diffs, and executes database migrations."

      api = container "API Gateway" {
        description "Exposes REST API for triggering and monitoring migrations"
        technology "Go (Gin), REST"
      }

      comparator = container "Schema Comparator" {
        description "Reads schemas from databases and identifies differences"
        technology "Go, SQL Introspection"
      }

      generator = container "Migration Generator" {
        description "Generates SQL migration scripts from schema diffs"
        technology "Go, SQL Builder"
      }

      executor = container "Migration Executor" {
        description "Applies generated migration scripts to the target database"
        technology "Go, SQL Executor"
      }

      tracker = container "Migration Tracker" {
        description "Tracks and logs migration history"
        technology "Go, Relational DB"
      }

      metadataDb = container "Metadata DB" {
        description "Stores migration history and metadata"
        technology "PostgreSQL"
      }

      sourceDb = container "Source Database" {
        description "The current (source) database to be compared"
        technology "Various SQL DBs"
      }

      targetDb = container "Target Database" {
        description "The database to apply migrations to"
        technology "Various SQL DBs"
      }

      user -> api "Triggers comparison/migration"
      cicd -> api "Triggers automated migration"

      api -> comparator "Request schema comparison"
      comparator -> sourceDb "Reads schema"
      comparator -> targetDb "Reads schema"
      comparator -> generator "Sends diff"

      generator -> executor "Sends SQL migration"
      executor -> targetDb "Applies migration"
      executor -> metadataDb "Records applied changes"

      api -> tracker "Check status/history"
      tracker -> metadataDb "Reads migration logs"
    }
  }

  views {
    systemContext system {
      include *
      autolayout lr
      title "System Context: Database Migration Software"
    }

    container system {
      include *
      autolayout lr
      title "Container Diagram: Database Migration Software"
    }

    theme default
  }
}
