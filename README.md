# Scaffold

## About

Scaffold is an infrastructure management tool that takes a waterfall approach to management. This works by tracking the state of each cascade (waterfall workflow DAG) task and triggering the next task as all its parents are in a success state. Cascades are manually triggered and report their state to the UI so engineers can track deployment and upgrade processes to ensure that they work as expected. Additionally, Scaffold allows for input and file storage and loading to make task execution easier. Finally, tasks are executed within container images to enable dependencies to be brought along to the task execution.

## Rationale

Infrastructure management is an interesting problem to tackle. Many organization use CI/CD pipelines to handle deployment and upgrades (e.g. Jenkins, Concourse) which allows for an automated approach to management, however execution status and can be hard to track. Additionally, changes that may be inside the middle of an upgrade pipeline may need to be re-run without running the whole pipeline and may change the required state of subsequent tasks. The waterfall approach allows for independent execution of tasks with their state changes propagating down to dependent tasks.

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

## TODO

### CLI

- **Authentication**
    - [ ] Generate API token from CLI
    - [ ] Read/write token from/to a `~/.scaffold/creds` file

- **Cascade Interaction**

    - [ ] Upload cascade
    - [ ] Download cascade
    - [ ] Delete cascade
    - [ ] Update cascade

- **Run Interaction**

    - [ ] Get run output
    - [ ] Get run statuses
        - [ ] By cascade name
        - [ ] By cascade and run names
    - [ ] Exec into run
    - [x] List available exec runs

- **File Interaction**

    - [ ] Upload a file
    - [ ] Download a file

- **Configuration**
    - [ ] Read/write config from/to a `~/.scaffold/config` file

### Server

-  **Dependency Interaction**

    - [x] Input changes set dependent tasks to `not_started`
    - [x] Run starts set dependent tasks to `not_started`

-  **Worker Improvements**

    - [x] Worker node directory and image cleanup
    - [ ] Task kill ability
    - [ ] Exec into finished container (if still around)
    - [x] Handle `no space left on device` if it happens

-  **Security Improvements**
    - [x] Hash and salt api tokens
    - [x] Hash and salt login tokens
    - [ ] Cascade group-based access
        - [ ] Files
        - [ ] Cascades
    - [ ] Encrypt exec websocket traffic
    - [ ] Run with HTTPS
    - [ ] Implement basic auth for API token request

-  **File UI**

    - [x] Files list page
    - [x] File upload
    - [x] File download

-  **UI Improvements**

    - [ ] Fix cascade search
    - [ ] Task search to highlight tasks containing search string
    - [x] Display Cascade links
    - [x] Display current Cascade state

-  **Task Display Improvements**

    - [x] Task store and show previous state
    - [ ] Task formatted display
        - Write specific format JSON to /tmp/run/.display in container to setup display in UI
            - [ ] Tables
            - [ ] Single value
            - [ ] Block value

-  **Documentation**

    - [ ] Setup readthedocs
    - [ ] Write documentation

-  **Cascade Improvements**

    - [x] Add check recurring tasks
    - [ ] Selective auto-execute
    - [ ] On success, on failure, and always tasks

-  **Manager Improvements**
    - [ ] Proxy websocket exec requests
    - [ ] Save worker proxy port on join
