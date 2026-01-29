<div align="center">
<img src=".github/resources/banner.png" width="auto" alt="Mortar wordmark">

![GitHub License][license-badge]
![GitHub Release][release-badge]
![GitHub Repo stars][stars-badge]
![GitHub Downloads][downloads-badge]

</div>

---

## How do I setup Pak Store?

Pak Store comes pre-installed with NextUI!

Simply connect your device to Wi-Fi and launch it from the `Tools` menu.

### Manual Installation

If for whatever reason you do not have the Pak Store:

1. Download the latest release from this repo.
2. Unzip the release download.
    - If the unzipped folder name is `Pak.Store.pak` please rename it to `Pak Store.pak`.
3. Copy the entire `Pak Store.pak` folder to `SD_ROOT/Tools/tg5040`.
4. Reinsert your SD Card into your device.
5. Launch `Pak Store` from the `Tools` menu and enjoy all the amazing Paks made by the community!

---

## I want my Pak in Pak Store!

Awesome! To get added to Pak Store you have to complete the following steps:

1. Create a `pak.json` file at the root of your repo. An example can be seen below.
    - The following fields are **required**
        - `name`
        - `version`
        - `type`
        - `description`
        - `author`
        - `repo_url`
        - `release_filename`
        - `platforms`
    - If you are packaging up an emulator, please set the name to the desired emulator tag. (e.g., an Intellivision Pak
      with the tag `INTV` would have `INTV` as the name in pak.json)
2. Prepare your Pak for distribution by making a zip file. The contents of the zip file must the contents present in the
   root of your Pak directory.
3. Ensure your release is tagged properly and matches the `version` field in `pak.json`.
    - The tag should be in the format `vX.X.X` where `X` is the major, minor, and patch version. For more details for
      using SemVer, please see the [SemVer Documentation](https://semver.org/).
    - 4-digit versions (`vX.X.X.X`) are also supported, but 3-digit is preferred.
    - GitHub releases have both tags and titles. The title does not matter in the context of the Pak Store but you
      should have it match the tag and pak.json version.
4. Make sure the file name of the release artifact matches what is in `pak.json`.
5. Once all of these steps are
   complete, [submit your pak using our issue form](https://github.com/LoveRetro/nextui-pak-store/issues/new?template=new-pak-submission.yml).
   You'll need to provide:
    - Your pak's display name (how it will appear in the store)
    - Your GitHub repository URL
    - The categories your pak belongs to
6. Someone will review your submission, may request changes, and then publish it.

---

## Sample pak.json

```json
{
  "name": "Pak Store",
  "version": "v3.0.0",
  "type": "TOOL",
  "description": "A Pak Store in this economy?!",
  "author": "The NextUI Community",
  "repo_url": "https://github.com/LoveRetro/nextui-pak-store",
  "release_filename": "Pak.Store.pak.zip",
  "changelog": {
    "v1.0.0": "Upgraded the UI to use gabagool, my NextUI Pak UI Library!"
  },
  "screenshots": [
    ".github/resources/screenshots/main_menu.jpg",
    ".github/resources/screenshots/browse.jpg",
    ".github/resources/screenshots/ports.jpg",
    ".github/resources/screenshots/portmaster_1.jpg",
    ".github/resources/screenshots/portmaster_2.jpg",
    ".github/resources/screenshots/updates.jpg"
  ],
  "platforms": [
    "tg5040",
    "tg5050"
  ]
}
```

---

# Community Shout Out!

Pak Store exists because of the incredible NextUI community. 

Your creativity, passion, and dedication to building amazing paks is what makes this platform special. 

Every emulator, tool, and enhancement you create brings joy to our retro doo-dads! 

Thank you for sharing your talents and making NextUI better for everyone! :heart:


<!-- Badge References -->

[license-badge]: https://img.shields.io/github/license/UncleJunVIP/nextui-pak-store?style=for-the-badge&color=9B2256

[release-badge]: https://img.shields.io/github/v/release/UncleJunVIP/nextui-pak-store?sort=semver&style=for-the-badge&color=9B2256

[stars-badge]: https://img.shields.io/github/stars/UncleJunVIP/nextui-pak-store?style=for-the-badge&color=9B2256

[downloads-badge]: https://img.shields.io/github/downloads/UncleJunVIP/nextui-pak-store/total?style=for-the-badge&label=Downloads&color=9B2256
