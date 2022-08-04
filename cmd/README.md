# /cmd

>This folder contains the main application entry point files for the project, with the directory name matching the 
>name for the binary. So for example `cmd/simple-service` meaning that the binary we publish will be `simple-service`.

From: https://www.wolfe.id.au/2020/03/10/how-do-i-structure-my-go-project/, accessed on 2022-08-01.

## /cmd/asteroid-api

This is the main entry point for the API.

## /cmd/keygen

This is a command line tool for generating a new keypair and signing nonce, since some tools across programming 
libs tend to be incompatible.
