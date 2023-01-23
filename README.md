# replace_image_tag
## Overview
- `go.mod` and `go.sum`: Describe the go environment.
- `init.go` and `src/`: Describe the source codes which updates the Manifest file of [k8s-argo repository](https://github.com/nayuta-ai/k8s-argo).
- `test/`: Describe the test codes corresponding to the source codes in `src` directory
## To Do
- [ ] Introduce the source codes in src directory to .github/workflows/update-docker.yaml for adding GitHub Actions
- [ ] Set up GitHub tag and align GitHub tag and Docker Image tag.
- [ ] Add GitHub Actions codes for src and test.
