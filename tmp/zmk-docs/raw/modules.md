# Modules

ZMK makes use of Zephyr modules to include additional source code or configuration files into its build. You can think of them as similar to plugins or themes. The most common uses of this feature are:

- Building firmware for a keyboard external to ZMK's tree
- Adding functionality to ZMK, such as a driver or a behavior

A common ZMK setup thus consists of the following separate components, commonly housed in their respective Git repositories.

## Building Locally

To add a module to your build when building locally, you will need to clone/copy said module into your local file tree. You can then build using the module as described in Building with External Modules.

## Beta Testing

You may find that there are some features which you desire for which there is a Pull Request, but no module. If this is the case, you can still make use of the feature.

### Developer Repositories and Branches

For a developer to submit a pull request to ZMK, they must first clone the original ZMK repository. After they have a copy of the source code, they may create a feature branch to work within. When they have finished, they will publish the feature branch and create the pull request.

#### Finding the repository page from the Pull Request

#### Finding the repository URL

#### Finding the repository branch

## Testing Features

### GitHub Actions

When using GitHub Actions to build ZMK, once you have obtained the correct URL, you'll need to modify the `west.yml` file similarly to Building With Modules. Add the remote for the branch like in said section. The difference is that you will need to change the selected remote and revision (or branch) for the `zmk` project.

#### Example

Default:
```
manifest:
  remotes:
    - name: zmkfirmware
      url-base: https://github.com/zmkfirmware
  projects:
    - name: zmk
      remote: zmkfirmware
      revision: main
      import: app/west.yml
  self:
    path: config
```

Alternative:
```
manifest:
  remotes:
    - name: zmkfirmware
      url-base: https://github.com/zmkfirmware
    - name: module_a_base
      url-base: https://github.com/user/module_a
  projects:
    - name: zmk
      remote: zmkfirmware
      revision: main
      import: app/west.yml
    - name: module_a
      remote: module_a_base
      revision: main
  self:
    path: config
```