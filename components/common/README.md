# Knative-Kafka (Shared) Source

Unlike the other "components" the code contained in this directory is NOT a standalone project and is NOT 
packaged / built into an executable deliverable.  Instead it is a collection of utilities that are used
by the various Knative-Kafka "component" projects.  The logic has been consolidated here to avoid excessive
duplication of code for obvious reasons.  You should only add code here when it is being used by multiple 
Knative-Kafka components, and not when you *think* it *might* end up being so.

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
