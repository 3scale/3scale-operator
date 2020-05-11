# Product Specification Reference

## Product conditions

A Product has a ProductStatus, which has an array of Conditions through which the Prodict has or has not passed. Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the Pod last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * Synced: the product has been synchronized with 3scale;
  * Orphan: the product spec contains reference(s) to non existing resources;
  * Invalid: the product spec is semantically wrong and has to be changed;
  * Failed: An error occurred during synchronization. The operator will retry.
