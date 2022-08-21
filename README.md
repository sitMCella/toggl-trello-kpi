# Key Perfomance Indicators (KPIs) for Toggl and Trello

![Grafana Dashboard](https://github.com/sitMCella/toggl-trello-kpi/wiki/images/Grafana_Dashboard_example.png)

## Table of contents

   * [Introduction](#introduction)
   * [Configuration](#configuration)
      * [Toggl Reports API](#toggl-reports-api)
      * [Trello API](#trello-api)
      * [Trello Cards](#trello-cards)
      * [Grafana](#grafana)
      * [Configure the Grafana plugins](#configure-the-grafana-plugins)
   * [Development](#development)
      * [Build project](#build-project)
      * [Run tests](#run-tests)
      * [Code format](#code-format)
   * [Run application](#run-application)
      * [Configure Toggl and Trello Data](#configure-toggl-and-trello-data)
      * [Run the Grafana Dashboard](#run-the-grafana-dashboard)
      * [PostgreSQL database client](#postgresql-database-client)

## Introduction

The following project provides a dashboard for Key Performance Indicators (KPIs) visualization. The data is retrieved from Trello and Toggl.

The dashboard is used for monitoring the tasks assigned to a specific team member on one or multiple Trello projects.

Grafana is used as visualization tool, and a custom dashboard has been created for visualizing the Toggl and Trello data.

Golang is used for retrieving the Toggl data using the Toggl Reports API and the Trello cards using the Trello API.

This project makes use of the Golang project https://github.com/adlio/trello.

Read the [application wiki](https://github.com/sitMCella/toggl-trello-kpi/wiki) for a detailed description of the application configuration and an usage example.

## Configuration

Create a file `configuration/settings.yml` using the file `configuration/settings_template.yml` as template.

The following sections describe how to fill the configuration file.

### Toggl Reports API

The Toggl API Token can be retrieved from the user profile in Toggl.

Add the Toggl API Token to the property "TOGGL_API_TOKEN" in `configuration/settings.yml`.

Run the following curl command for testing purposes. The command retrieves the user workspaces in Toggl.

```
curl -v -u {Toggl_API_Token}:api_token -H "accept: application/json" -X GET https://api.track.toggl.com/api/v8/workspaces | jq '.'
```

### Trello API

Generate the Trello App Key from the web page:
 `https://trello.com/app-key`

Add the Trello App Key to the property "TRELLO_APP_KEY" in `configuration/settings.yml`.

Generate the API Token for the application from the following web page.

```sh
https://trello.com/1/authorize?key=<Trello_App_Key>&name=my_api_key&expiration=30days&response_type=token
```

Add the Trello API Token to the property "TRELLO_API_TOKEN" in `configuration/settings.yml`.

Run the following curl command to test the REST API call.

```sh
curl -H "accept: application/json" -X GET "https://api.trello.com/1/members/me/boards?key={Trello_App_Key}&token={Trello_Token}" | jq '.'
```

Run the following curl command in order to retrieve the list of boards from Trello.

```sh
curl -H "accept: application/json" -X GET "https://api.trello.com/1/members/me/boards?fields=name,url&key={Trello_App_Key}&token={Trello_Token}" | jq '.'
```

Extract the ID of the Trello board from the output of the last curl command.

Add the board ID to the property "TRELLO_BOARD_ID" in `configuration/settings.yml`.

Run the following curl command in order to retrieve the labels configuration of a specific board from Trello.

```sh
curl -H "accept: application/json" -X GET "https://api.trello.com/1/boards/{Board_ID}/labels?key={Trello_App_Key}&token={Trello_Token}" | jq '.'
```

### Trello Cards

Each card in Trello should contain 4 label types:
 - Project
 - Customer
 - Team
 - Type

The "Project" labels define the project names.
The "Customer" labels define the customer names.
The "Team" labels define the team names.
The "Type" labels define the different tasks, e.g. "Design", or "Implementation".

Each one of the label type should contain a range of labels.
The labels that corresponds to the same label type should have the same color, or a finite number of colors.

Add the label colors to the corresponding properties in `configuration/settings.yml`. For example:

```yaml
TRELLO_LABEL_PROJECT_COLOR: ["sky"]
TRELLO_LABEL_CUSTOMER_COLOR: ["green"]
TRELLO_LABEL_TEAM_COLOR: ["yellow"]
TRELLO_LABEL_CARD_TYPE_COLOR: ["red", "blue"]
```

### Grafana

Grafana is used as the visualization tool for the Toggl and Trello data.

The Grafana dashboard is configured to visualize time set data in a specific time range.

Configure the Grafana properties in `configuration/settings.yml`. For example:

```yaml
GRAFANA_YEAR: "2021"
GRAFANA_START_MONTH: "02"
GRAFANA_END_MONTH: "03"
```

Make sure to run the application with the choice_id 8 (see below), in order to complete the configuration of the Grafana Dashboard.

### Configure the Grafana plugins

This project makes use of the Grafana plugin: https://github.com/grafana/piechart-panel.git.

Run the following commands to install the plugin:

```sh
cd ./grafana/plugins
git clone https://github.com/grafana/piechart-panel.git --branch release-1.3.8
```

## Development

Prerequisites:
* Golang 1.19+
* Docker
* Docker compose

### Build project

Use the proper GOOS and GOARCH parameters from https://golang.org/doc/install/source#environment.

```sh
env GOOS=[host_operating_system] GOARCH=[host_cpu] go build -o toggl-trello-kpi cmd/main.go
```

### Run tests

```sh
go test ./...
```

### Code format

Use gofmt to format the source code:

```sh
gofmt -s -w .
```

## Run application

The project consists of a binary application used for retrieving the data from Trello and Toggl using the REST APIs, and a Docker compose application for the dashboard.

Docker compose is used for running the database and Grafana.

The usual flow is the following: 1) Create the Grafana dashboard, 2) Retrieve and configure Toggl and Trello data, 3) Run the Grafana dashboard, 4) Send the data to the database.

### Configure Toggl and Trello Data

Run the application:
```sh
./toggl-trello-kpi -choice={choice_id} [options]
```

Where:
 - choice_id: 1 -> Download the Toggl Time data as CSV file.
 - choice_id: 2 -> Download the Trello cards as CSV file.
 - choice_id: 3 -> Insert either the Toggl Time entries or the Trello card entries into the database from a CSV file.
 - choice_id: 8 -> Create the Grafana dashboard.

Example 1. Download the Toggl Time data as CSV file:
 `./toggl-trello-kpi -choice=1 2021 02`

Example 2. Download the Trello cards as CSV file:
 `./toggl-trello-kpi -choice=2`

Example 3.1. Insert the Toggl Time data into the database from a CSV file:
 `./toggl-trello-kpi -choice=3 toggl_time_entries.csv toggl_time`

Example 3.2. Insert the Trello cards into the database from a CSV file:
 `./toggl-trello-kpi -choice=3 trello_entries.csv trello_card`

Example 8. Create the Grafana Dashboard from the configuration defined in "configuration/settings.yml":
  `./toggl-trello-kpi -choice=8`

The choices 4, 5, 6 and 7 provide optional features.

### Run the Grafana Dashboard

Run the following command.

```sh
docker-compose -f docker-compose.yml up
```

### PostgreSQL database client

Run the PostgreSQL client (requires psql).

```sh
psql -h 127.0.0.1 -p 5432 -U postgres -W -d toggltrelloapi
```
