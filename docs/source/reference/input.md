# Input

Inputs allow you to have control over configuration of your workflow without having to re-deploy the workflow each time you need to change something. These values will be read in by the relevant tasks as environment variables

Inputs of type `secret` will have their values masked in the UI

## Schema

```yaml
name: str # input name
description: str # input label text
default: str # default value
type: str # input type. `secret|plaintext`
```
