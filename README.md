# atlas-sync-action

A GitHub Action for syncing a migration directory to [Atlas Cloud](https://atlasgo.cloud).

## Usage

Use this action in your workflow to sync your migration directory to Atlas Cloud.
It is highly recommended that you only run this action on the main branch of your repository,
as its purpose is to persist the desired state of your migration directory.

To use this action, you must first have a token with permissions to write to your
account. To create it:
- Login to [Atlas Cloud](https://atlasgo.cloud)
- Navigate to the Settings page and select the Bots tab. 
- Create a new bot and then create a new token for it, copy the token.
- Add this token as a secret to your repository, named `ATLAS_CLOUD_TOKEN`.

Then, in your workflow, add the following step:

```yaml
name: Sync Atlas Migrations
on:
  push:
    branches:
      - master # Only run on our main branch
jobs:
  sync:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8
        env:
          MYSQL_DATABASE: dev
          MYSQL_ROOT_PASSWORD: pass
        options: >-
          --health-cmd "mysqladmin ping -ppass"
          --health-interval 10s
          --health-start-period 10s
          --health-timeout 5s
          --health-retries 10
    steps:
      - uses: actions/checkout@v3
      - uses: ariga/atlas-sync-action@v0
        with:
          dir: 'migrations'
          name: 'app'
          dev-url: 'mysql://root:pass@mysql:3306/dev'
          cloud-token: ${{ secrets.ATLAS_CLOUD_TOKEN }}
```

## Configuration

Configure this action with the following inputs:

#### `dir`

The path to the directory containing your migrations.

#### `name` (optional)

The name of the migration directory in Atlas Cloud.

#### `tag` (optional)

When continuously syncing your directory, this input can be used to provide a unique identifier for each version. Defaults to the commit SHA.

#### `dev-url`

The URL of the dev database to connect to. Atlas will use this database to check the validity of the SQL files before syncing them to Atlas Cloud.

#### `cloud-token`

The Atlas Cloud token to use for authentication. Must be passed as a secret.

## Next steps

1. [`ariga/atlas-action`](https://github.com/ariga/atlas-action) - A GitHub Action to set up CI for database schema changes.
2. [`ariga/atlas-deploy-action`](https://github.com/ariga/atlas-deploy-action) - A GitHub Action to deploy schema migrations.
