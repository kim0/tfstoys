# Terraform State Toys

This is a collection (well currently only diff exists) of tools/toys to work with terraform state stored on S3. This is mostly customized for my use-case and not too generic.

## Usage

```
go install github.com/kim0/tfstoys
tfstoys diff --help
```

If you don't supply arguments, you get prompted to interactively select a state object and pick versions to compare
