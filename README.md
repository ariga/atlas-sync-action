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
    steps:
      - uses: actions/checkout@v3
      - uses: ariga/atlas-sync-action@v0
          with:
            dir: path/to/migrations
            name: my_app
            tag: ${{ github.sha }} # optional
            dev-url: "sqlite://file?cache=shared&mode=memory"
            cloud-token: ${{ secrets.ATLAS_CLOUD_TOKEN }}
```

## Configuration

Configure this action with the following inputs:

#### `dir`

The path to the directory containing your migrations.

#### `name`

The name of the migration directory in Atlas Cloud.

#### `tag` (optional)

The tag of the specific version of the migration directory.


#### `dev-url` (optional)

The URL of the dev database to connect to.

#### `cloud-token`

The Atlas Cloud token to use for authentication. Must be passed as a secret.