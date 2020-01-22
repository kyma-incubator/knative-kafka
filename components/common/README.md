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

## Updating The Common Component Instructions

In order to update a consumer of this **common** component with changes made here, you need to first have pushed a 
branch with those changes to your fork. Next, test your changes in the consuming project by updating the 
dependency on common in the respective Gopkg.toml to follow the pattern
```toml
[[constraint]]
   name = "github.com/kyma-incubator/knative-kafka"
   branch = "<branch-name-in-fork-repo>"
   source = "github.com/<fork-repo>/knative-kafka"
```
Then update the vendor directory by running
```bash
dep ensure -update github.com/kyma-incubator/knative-kafka
```

Once you have thoroughly tested the changes of the common component, submit a PR for **ONLY** the common component 
changes. After this PR is approved and the changes are on the master branch, revert the dependency in the 
Gopkg.toml of the consuming project to once again point to the master branch of the kyma-incubator/knative-kafka 
repo. 

Finally, go thru the consuming components (channel, controller, dispatcher) and update the common dependency via
```bash
dep ensure -update github.com/kyma-incubator/knative-kafka
```
These updates will be a part of the final PR that includes any changes made as part of feature/bug work on the 
knative-kafka components. If the feature/bug changes were only in common, this second PR is still required since 
the changes to common must be on the master branch before dep can pull in the changes.
