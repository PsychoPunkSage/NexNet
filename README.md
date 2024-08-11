# NexNet

## Installation:

### Setting up your development environment

`NexNet` is written in GO. Before you run the code, you need to set up your GO development environment.

1. Install `Go` version **1.16** or above.
2. Define `GOPATH` environment variable and modify `PATH` to access your Go binaries. A common setup is as follows. You could always specify it based on your own flavor.

    ```sh
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    ```

### Step 1: Fork in the cloud

1. Visit https://github.com/PsychoPunkSage/NexNet
2. On the top right of the page, click the `Fork` button (top right) to create
   a cloud-based fork of the repository.

### Step 2: Clone fork to local storage
Create your clone:

```sh
mkdir -p $working_dir
cd $working_dir
git clone https://github.com/$user/NexNet.git
# or: git clone git@github.com:$user/NexNet.git

cd $working_dir/NexNet
git remote add upstream https://github.com/PsychoPunkSage/NexNet.git
# or: git remote add upstream git@github.com:PsychoPunkSage/NexNet.git

# Never push to the upstream master.
git remote set-url --push upstream no_push

# Confirm that your remotes make sense:
# It should look like:
# origin    git@github.com:$(user)/NexNet.git (fetch)
# origin    git@github.com:$(user)/NexNet.git (push)
# upstream  https://github.com/PsychoPunkSage/NexNet (fetch)
# upstream  no_push (push)
git remote -v
```
