# Scaffold

## About

Scaffold is an infrastructure management tool that takes a waterfall approach to management. This works by tracking the state of each cascade (waterfall workflow DAG) task and triggering the next task as all its parents are in a success state. Cascades are manually triggered and report their state to the UI so engineers can track deployment and upgrade processes to ensure that they work as expected. Additionally, Scaffold allows for input and file storage and loading to make task execution easier. Finally, tasks are executed within container images to enable dependencies to be brought along to the task execution.

## Rationale

Infrastructure management is an interesting problem to tackle. Many organization use CI/CD pipelines to handle deployment and upgrades (e.g. Jenkins, Concourse) which allows for an automated approach to management, however execution status and state can be hard to track. Additionally, changes that may be inside the middle of an upgrade pipeline may need to be re-run without running the whole pipeline and may change the required state of subsequent tasks. The waterfall approach allows for independent execution of tasks with their state changes propagating down to dependent tasks.

## Examples

- Simple
    ```yaml
    version: v1
    name: foobar
    inputs:
    - name: file_contents
        description: File contents to read and write
        type: text
        default: "foobar"
    - name: message_contents
        description: Message contents to write
        type: password
        default: "hello world"
    tasks:
    - name: write_file
    image: ubuntu:20.04
    run: |
        echo "FOOBAR" > /tmp/run/myFile.txt
    store:
        file:
        - myFile.txt
    - name: store_env_var
    image: ubuntu:20.04 
    run: |
        MESSAGE="Hello world"
    store:
        env:
        - MESSAGE
    - name: print_file_and_env
    depends_on:
        - write_file
        - store_env_var
    image: ubuntu:20.04
    load:
        env:
        - MESSAGE
        file:
        - myFile.txt
    run: |
        ls /tmp/run
        cat /tmp/run/myFile.txt
        echo "${MESSAGE}"
    ```

- EKS Application deployment
    ```yaml
    version: v1
    name: application-deploy
    inputs:
    - name: ARTIFACTORY_USERNAME
        description: Artifactory username
        type: text
        default: "myCoolArtifactoryUsername"
    - name: ARTIFACTORY_PASSWORD
        description: Artifactory password
        type: password
        default: "myCoolArtifactoryPassword"
    - name: POSTGRES_ADMIN_USERNAME
        description: application API Postgres username 
        type: text
        default: "myCoolAdminPostgresUsername"
    - name: POSTGRES_ADMIN_PASSWORD
        description: application API Postgres password
        type: password
        default: "myCoolAdminPostgresPassword"
    - name: POSTGRES_PORTAL_USERNAME
        description: application Portal Postgres username
        type: text
        default: "myCoolPortalPostgresUsername"
    - name: POSTGRES_PORTAL_PASSWORD
        description: application Portal Postgres password
        type: password
        default: "myCoolArtifactoryPassword"
    - name: POSTGRES_API_USERNAME
        description: application API Postgres username 
        type: text
        default: "myCoolArtifactoryPassword"
    - name: POSTGRES_API_PASSWORD
        description: application API Postgres password
        type: password
        default: "myCoolArtifactoryPassword"
    tasks:
    - name: deploy_terraform
    image: ubuntu:20.04
    run: |
        echo "FOOBAR" > /tmp/run/tfstate
        sleep 1
        echo "Terraform deployed"
    store:
        file:
        - tfstate
    - name: create_db_schemas
    image: ubuntu:20.04
    run: |
        echo "Admin username: ${POSTGRES_ADMIN_USERNAME}"
        echo "Admin password: ${POSTGRES_ADMIN_PASSWORD}"
        sleep 1
        echo "Schemas created"
    inputs:
        POSTGRES_ADMIN_USERNAME: POSTGRES_ADMIN_USERNAME
        POSTGRES_ADMIN_PASSWORD: POSTGRES_ADMIN_PASSWORD
    depends_on:
    - deploy_terraform
    - name: create_db_users
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "Users created"
    inputs:
        POSTGRES_API_USERNAME: POSTGRES_API_USERNAME
        POSTGRES_API_PASSWORD: POSTGRES_API_PASSWORD
        POSTGRES_PORTAL_USERNAME: POSTGRES_PORTAL_USERNAME
        POSTGRES_PORTAL_PASSWORD: POSTGRES_PORTAL_PASSWORD
    depends_on:
        - create_db_schemas
    - name: build_images
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "Images built"
    inputs:
        ARTIFACTORY_USERNAME: ARTIFACTORY_USERNAME
        ARTIFACTORY_PASSWORD: ARTIFACTORY_PASSWORD
    - name: push_images
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "Images pushed"
    inputs:
        ARTIFACTORY_USERNAME: ARTIFACTORY_USERNAME
        ARTIFACTORY_PASSWORD: ARTIFACTORY_PASSWORD
    depends_on:
    - build_images
    - name: deploy_application
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "application deployed"
    depends_on:
    - create_db_users
    - push_images
    - name: deploy_traefik
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "Traefik deployed"
    depends_on:
    - create_db_users
    - push_images
    - name: run_tests
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "Tests passed"
    depends_on:
    - deploy_application
    - name: register_new
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "New instance registered"
    depends_on:
    - run_tests
    - deploy_traefik
    - name: east_coast_check
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "East coast can reach instance"
    depends_on:
    - register_new
    - name: midwest_check
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "East coast can reach instance"
    depends_on:
    - register_new
    - name: west_coast_check
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "West coast can reach instance"
    depends_on:
    - register_new
    - name: deregister_old
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "Old instance de-registered"
    depends_on:
    - east_coast_check
    - midwest_check
    - west_coast_check
    - name: destroy_old_application
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "Old application destroyed"
    depends_on:
    - deregister_old
    - name: destroy_old_traefik
    image: ubuntu:20.04
    run: |
        sleep 1
        echo "Old Traefik destroyed"
    depends_on:
    - deregister_old
    ```

## ROADMAP

### v0.1.0

- **CLI**

    - __Miscellaneous__

        - [ ] Log level configuration
        - [x] CLI version command
        - [x] CLI get Scaffold version

    - __Authentication__

        - [x] Generate API token from CLI
        - [x] Read/write token from/to a `~/.scaffold/creds` file

    - __Cascade Interaction__

        - [x] Upload cascade
        - [x] Delete cascade
        - [x] Update cascade
        - [x] Set cascade context in profile

    - __Object Interaction__

        - [x] Get cascades
        - [x] Get cascade by name
        - [x] Describe cascade
        - [x] Get datastores
        - [x] Get datastore by name
        - [x] Describe datastore
        - [x] Get states
        - [x] Get state by name
        - [x] Describe state
        - [x] Get tasks
        - [x] Get task by name
        - [x] Describe task

    - __RUN Interaction__

        - [x] Exec into run
        - [x] List available exec runs

    - __File Interaction__

        - [x] Upload a file
        - [x] Download a file

    - __Configuration__

        - [x] Read/write config from/to a `~/.scaffold/config` file

- **Server**

    -  __Dependency Interaction__

        - [x] Input changes set dependent tasks to `not_started`
        - [x] Run starts set dependent tasks to `not_started`

    -  __Worker Improvements__

        - [x] Worker node directory and image cleanup
        - [x] Task kill ability
        - [x] Exec into finished container (if still around)
        - [x] Handle `no space left on device` if it happens

    -  __Security Improvements__
        - [x] Hash and salt api tokens
        - [x] Hash and salt login tokens
        - [x] Cascade group-based access
            - [x] Files
            - [x] Cascades
        - [x] Run with HTTPS
        - [x] Implement basic auth for API token request
        - [x] Group based access for container exec

    -  __File UI__

        - [x] Files list page
        - [x] File upload
        - [x] File download

    -  __UI Improvements__

        - [x] Fix cascade search
        - [x] Task search to highlight tasks containing search string
        - [x] Display Cascade links
        - [x] Display current Cascade state
        - [x] Add group and role removal from user create page
        - [x] Add legend to cascade page
        - [x] Fix user search
        - [x] Task store and show previous state
        - [x] Task formatted display
            - Write specific format JSON to /tmp/run/.display in container to setup display in UI
                - [x] Tables
                - [x] Single value
                - [x] Pre-formatted value
        - [x] Previous should display one less than run number
        - [x] Fix status page to show more than just healthy nodes

    -  __Documentation__

        - [ ] Setup readthedocs
        - [ ] Write documentation
        - [x] Swagger docs

    -  __Cascade Improvements__

        - [x] Add check recurring tasks
        - [x] Selective auto-execute
        - [x] On success, on error, and always tasks

    -  __Manager Improvements__

        - [x] Proxy websocket exec requests
        - [x] Save worker proxy port on join
        - [x] Serve version endpoint
    
    - __Testing__

        - [x] Basic CLI configure interaction testing
        - [x] Basic CLI apply interaction testing
        - [x] Basic CLI get interaction testing
        - [x] Group-based access testing
        - [x] Run trigger testing
        - [x] Cascade REST testing
        - [x] DataStore REST testing
        - [x] Input REST testing
        - [x] State REST testing
        - [x] Task REST testing
        - [x] >= 50% test coverage

    - __Logging__
        
        - [ ] Configure log format

### v0.2.0

- **Server**
    - __Testing__

        - [ ] 75% test coverage

    -  __UI Improvements__

        - [ ] Move task status to right slide in
        - [ ] Change auto-execute from cascade page

    -  __Worker Improvements__

        - [ ] Check task kill ability
