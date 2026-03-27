# Context Terraform Provider

The `context` provider for Terraform and OpenTofu is designed to solve a common, complex problem: generating **consistent, predictable, and standardized names and tags** for all your resources.

While many solutions offer simple string formatting, the `context` provider's key differentiator is its power and flexibility. It is **fully customizable using JQ expressions**. This unique approach allows you to define sophisticated, conditional logic for naming and tagging, much like writing a custom function. You can create rules that adapt based on the environment, component, or any other metadata you provide.

## Core Concept

The provider works by building a single, comprehensive **"context object"** (a structured JSON document). This object acts as a central repository for all the metadata about your infrastructure stack.

Here's the workflow:

1. **Collect:** As Terraform walks through your modules (from root to child components), the provider's components collect metadata (like project, environment, application name, etc.) and progressively merge it into the central context object.
2. **Execute:** This `item_module` takes the complete, fully-assembled context object and executes your custom **JQ expression** against it.
3. **Generate:** The JQ script maps, filters, and transforms the data to produce the final, desired **name** and **tags** as outputs, which you can then pass directly to your resource.

## Key Components

The provider is built around a few key elements that work together to build and evaluate the context:

* **`namespace`:**
  This element logically groups a set of related resources or components. It's perfect for adding shared metadata to a specific subset of your infrastructure, such as everything belonging to a particular team or microservice.

* **`label`:**
  This is used for each resource you want to name. It is the final step that triggers the evaluation. It takes the fully assembled context object, runs the configured JQ expression, and provides the final **`id`** and **`tags`** as outputs.

* **`variables`:**
  A helper data source used to extract all variables and their evaluated values from a context stack.

## Requirements

* [Mise](https://mise.jdx.dev/getting-started.html) >= 2025.11.14: Mise will install the necessary tools

## Building and Using The Provider

1. Clone the repository
2. Enter the repository directory
3. Install the [Required tools](#requirements)
4. Build the provider using the Mise command:

```shell
mise run install
```

5. Navigate to `examples` folder and choose an example tf file
6. Run the following Mise command inside the chosen example directory to plan the Terraform/OpenTofu code:

```shell
mise run tf-plan
```

## Developing the Provider

This section outlines the preconfigured mise commands that streamline the provider's development workflow. These commands cover essential tasks such as managing dependencies, running tests, and generating documentation.

### Adding and Clearing Dependencies

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
mise run add-dependency github.com/author/dependency
```

To clean up and synchronize the project's dependency files (go.mod and go.sum) with your actual source code run the following command:

```shell
mise run clean-dependencies
```

### Run Acceptance Tests

To run all of the tests for the provider execute the following command:

```shell
mise run test-all
```

To only run a specific test execute the following command:

```shell
mise run test <test function name> --path <relative path to the test directory>
```

If you in the directory where the test is located the  `--path` flag can be skipped:

```shell
mise run test <test function name>
```

### Generate documentation

To generate or update documentation, run:

```shell
mise run generate-doc
```

## Releases

To release a new version of the provider, you have to create a tag locally, annotated and signed, from the main branch, following semver format, and push it to GitHub.  
GitHub Actions will trigger the release workflow that will create a release in GitHub from the tag, that will include the provider built binaries.  
**NOTE**: GitHub UI cannot create tags without publishing a release, and since releases are immutable, the pipeline that will be triggered will not be able to add the built binaries to the release.
