# Info

This module will be used by the controllers in the Extension Manager Operator; and also by the Extension Manager Service, 
where this module will be used to validate the local JSON before being fed to Jukebox (so developers can receive the same 
errors locally than they'd receive once the ContentConfiguration is deployed on-cluster).

Requirements:

Create a package in the `golang-commons` library called ContentConfiguration.

The package will contain the following features:

validate input configuration as JSON or YAML for correctness
Generate configurationResult that will be stored in ContentConfiguration.status.configurationResult
Return descriptive error messages in case validation fails
Support inlineConfiguration
Implemente schema validation for the input configuration
