# sample_golang_app

Sample app written in Go for connecting to the DigitalOcean SaaS marketplace

## Running Locally

To start the database locally, run `docker compose up`. Once running, it will accept connections on port 5431, and a UI can be found at localhost:8080.

To start the server locally, simply use `make run`. It will then be available on localhost:8082.

## To Use

This is intended to be a starting point for anyone looking to write a DigitalOcean SaaS Add-on. It contains endpoints for all calls DigitalOcean will make to a SaaS Add-on, as well as a couple of endpoints intended for use by a front-end to call back to DigitalOcean for configuration changes. If you want to use this, you will likely find the files under `/internal/server` to be the most helpful.

If you want to run this as it is, consider using DigitalOcean's [App Platform](https://www.digitalocean.com/go/app-platform?utm_campaign=amer_brand_kw_en_cpc&utm_adgroup=digitalocean_app_platform_exact&_keyword=digital%20ocean%20app%20platform&_device=c&_adposition=&utm_content=conversion&utm_medium=cpc&utm_source=google&gclid=CjwKCAjw2OiaBhBSEiwAh2ZSP4ZmQPsVuzTJh-AZj-RpancsW5YvXbjAitPG_FTHgpmymtvUro7j7RoCiwoQAvD_BwE).

## Database Tables

This app assumes a database exists containing three tables: Accounts, Activiites, and Tokens. Accounts represent the user accounts on your system, also referred to as Resources. Activiites represent an audit log of actions taken - in this example, all Notifications sent to the Add-on are written here. Tokens represent oauth grants.

For additional details, see `init.sql` or the provided UI as detailed in **Running Locally**.

### Accounts

| Column           | Type                   |
|------------------|------------------------|
| id               | integer Auto Increment |
| name             | character varying      |
| email            | character varying      |
| app_slug         | character varying NULL |
| plan_slug        | character varying NULL |
| resource_uuid    | character varying      |
| language         | character varying      |
| email_preference | boolean                |
| source           | character varying NULL |
| source_id        | character varying NULL |
| status           | smallint               |
| license_key      | character varying      |
| created_at       | timestamptz            |
| modified_at      | timestamptz            |

### Activiites

| Column        | Type                   |
|---------------|------------------------|
| id            | integer Auto Increment |
| account_id    | integer                |
| resource_uuid | character varying NULL |
| type          | character varying      |
| title         | character varying      |
| body          | character varying      |
| created_at    | timestamptz            |
| modified_at   | timestamptz            |

### Tokens

| Column        | Type                    |
|---------------|-------------------------|
| id            | integer Auto Increment |
| resource_uuid | character varying       |
| access_token  | character varying       |
| refresh_token | character varying       |
| expires_at    | timestamptz             |
| issued_at     | timestamptz             |


## Further Documentation

For additional details on the API DigitalOcean expects from its Add-ons, go [here](https://marketplace.digitalocean.com/vendors/saas-api-docs).

This app was designed to work with a single-page application built in React, an example of which can be found [here](https://github.internal.digitalocean.com/oadesokan/starter-app-nodejs).