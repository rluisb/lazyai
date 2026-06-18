# Documentation

> danger
> 
> If any of the above steps throw an error, they need to be addressed and all of the checks re-run prior to submitting a pull request.

> note
> 
> The documentation uses American English spelling and grammar conventions. Title case is used for the first three heading levels, with sentence case used beyond that.
> 
> Please make sure your changes conform to these conventions - prettier and lint are unfortunately unable to do this automatically.

## Submitting a Pull Request

Once the above sections are complete the documentation updates are ready to submit as a pull request.

## Formatting and Linting Your Changes

Prior to submitting a documentation pull request, you'll want to run the format and check commands. These commands are run as part of the verification process on pull requests so it's good to run them ahead of submitting documentation updates.

The format command can be run with the following procedure in a terminal that's inside the ZMK source directory.

1. Navigate to the `docs` folder
2. Run `npm run prettier:format`

The check commands can be run with the following procedure in a terminal that's inside the ZMK source directory.

1. Navigate to the `docs` folder
2. Run `npm run prettier:check`
3. Run `npm run lint`
4. Run `npm run build`

> danger
> 
> If you are working with the documentation from within VS Code+Docker please be aware the documentation will not be auto-generated when making changes while the server is running. You'll need to restart the server when saving changes to the documentation.

> note
> 
> You will need `Node.js` and `npm` installed to update the documentation. If you're using the ZMK dev container (Docker) the necessary dependencies are already installed. Otherwise, you must install these dependencies yourself. Since `Node.js` packages in Linux distributions tend to be outdated, it's recommended to install the current version from a repository like NodeSource to avoid build errors.

## Testing Documentation Updates Locally