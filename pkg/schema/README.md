The architecture of this package is heavily influenced by [Skaffold](https://github.com/GoogleContainerTools/skaffold).
The intent of the schema package is to allow for versioning and upgrading of stack configs.
As new features are added to stack, config will change, and this can lead to cases where newer versions of stack no longer
work for previous versions. 

Config files for stack should contain an `ApiVersion` of the form `stack/{version}{release}`.  
The earliest versions of stack do not have an ApiVersion, so files like this are treated as `stack/v0beta1`.
The current version is `stack/v1alpha1`. From a given version, the schema tooling will attempt to 
upgrade that file to the latest schema. In some cases, versions may not be compatible for upgrade, and upgrade jobs will report back as such. 
In these cases, users may manually upgrade their configuration, or install an older version of stack.



