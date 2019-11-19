# Knative-Kafka (Shared) Source

Unlike the other "components" the code contained in this directory is NOT a standalone project and is NOT 
packaged / built into an executable deliverable.  Instead it is a collection of utilities that are used
by the various Knative-Kafka "component" projects.  The logic has been consolidated here to avoid excessive
duplication of code for obvious reasons.  You should only add code here when it is being used by multiple 
Knative-Kafka components, and not when you *think* it *might* end up being so.

## Setup
The following are initial project setup instructions above and beyond those in parent [README](../README.md)...
```
# Setup The Local Dev Environment Variables (Do this for every new shell / session)
. ./local-env.sh
```


## Makefile
The Makefile should be used to build and test the code as follows...
    
- **Dependencies**
    ```
    # Verify / Download Dependencies
    make dep
    ```    
    
- **Test**
    ```
    # Run All Unit Tests
    make test
    ```

## Update Instructions

In order to update a consumer of this **common** package with changes made here, you need to first have pushed a 
PR with those changes through to the develop/ branch.  Then you need to change to the consuming project and update
the dep dependency as follows...
```
dep ensure -update github.com/kyma-incubator/knative-kafka
```
