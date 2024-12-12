# Hello World

## Creating a Workflow

Now that you have Scaffold up and running and your CLI configured, let's build a basic workflow to take you through working with Scaffold

Scaffold workflows use inputs and tasks to define the execution steps

- Inputs are objects which allow you to setup ui-managed configuration values. For example, you may want to clone a specific branch so you might create a `branch` input that defaults to `main`, but allows you to change the branch you're going to pull from as needed

- Tasks are the actual execution steps which make up the workflow logic

To get started on your workflow, we first need to setup the basic structure

```yaml
version: v1
name: hello_world
groups: 
  - foo
inputs: []
tasks: []
```

In this example, we've created an empty workflow with the name `hello_world` which the `foo` group has access too

## Adding an Input

Now that we have our workflow, let's add an input to allow us to control who we're saying hello to

```yaml
version: v1
name: hello_world
groups:
- foo
inputs:
- name: name
  description: "Who to say hello to"
  default: "World"
  type: plaintext
tasks: []
```

Now we have the ability to change who we say hello to, with `World` being the default name

## Adding tasks

So we've created an input, now we just need to setup the tasks to actually use it and do some work

Let's add a task to take in that input and craft our greeting message

```yaml
version: v1
name: hello_world
groups:
- foo
inputs:
- name: greeting_name
  description: "Who to say hello to"
  default: "World"
  type: plaintext
tasks:
- name: setup_message
  kind: local
  run: |
    export message="Hello ${greeting_name}"
    echo "message stored"
  store:
    env:
      - message
  inputs:
    greeting_name: greeting_name
```

We've now added our first task, which we can breakdown as follows:

Our task is named as `setup_message` and is used to create the greeting we will display. We are running locally on the worker, no need to deal with containers at this point. Additionally, we will load in the value from the `greeting_name` input we created earlier and export the environment variable `message` to be equal to `Hello ${greeting_name}`
> NOTE: To store environment variable values they _must_ be exported
After doing this, we store the message variable to our context to be used later

So we have one task, but it just stores our message it doesn't really do anything with it, so let's add another task to change that

```yaml
version: v1
name: hello_world_2
groups:
- foo
inputs:
- name: greeting_name
  description: "Who to say hello to"
  default: "World"
  type: plaintext
tasks:
- name: setup_message
  kind: local
  run: |
    export message="Hello ${greeting_name}"
    echo "message stored"
  store:
    env:
      - message
  inputs:
    greeting_name: greeting_name
- name: echo_message
  kind: local
  auto_execute: true
  depends_on:
    success:
      - setup_message
  run: |
    echo "${message}"
  load:
    env:
      - message
```

Here we've added our second step with similar configuration but with a few key differences:

- `echo_message` depends on the success of the `setup_message` and with auto-execute enabled it will run once `setup_message` passes
- We don't persist anything to our context, instead we're loading in the `message` variable from it
- We don't use any inputs, since that's already been used in the `setup_message` step there's no point

Save this workflow at `hello-world.yaml` and we can move on to the next step

## Applying The Workflow

We're written a workflow with two tasks, so now what?

Next we need to apply our workflow so that we can use it in Scaffold:

```sh
scaffold apply workflow -f hello-world.yaml
```

Once this has been applied, go to [http://localhost:2997](http://localhost:2997) and login with the username `admin` and password `admin`

Once you're logged in, click on the link icon corresponding to our workflow to take you to the workflow page

Next, double click the `setup_message` step and click on the play button on the top right of the box that appears

Close out of that box and you should see the `setup_message` step either running (blue) or finished (green)

Once it finishes, you'll see the `echo_message` step run automatically. Double click its box and you can expand the `context` section to view the stored context that was passed down from the previous step. Additionally, view the `output` section to see your message

So what next? Try clicking to the `Inputs` button on the top right and changing who you want to say hello to. Click on the save icon and you can now re-run your `setup_message` step to see the greeting change to whoever you decided
