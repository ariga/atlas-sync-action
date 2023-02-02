# atlas-sync-action

A GitHub Action for syncing a migration directory to [Atlas Cloud](https://atlasgo.cloud).

## Usage

Use this action in your workflow to sync your migration directory to Atlas Cloud.
It is highly recommended that you only run this action on the main branch of your repository,
as it's purpose is to persist the desired state of your migration directory.

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
  gen:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: atlasgo/atlas-sync-action@v0
          with:
            dir: path/to/migrations
            driver: mysql # or: postgres | sqlite
            cloud-token: ${{ secrets.ATLAS_CLOUD_TOKEN }}
```

## Configuration

Configure this action with the following inputs:

#### `dir`

**Required** The path to the directory containing your migrations.

#### `driver`

**Required** The database driver to use. One of: `mysql`, `postgres`, `sqlite`.

#### `cloud-token`

**Required** The Atlas Cloud token to use for authentication.

The full list of options can be found in the [action.yml](action.yml) file.