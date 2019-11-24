---
title: "Translations"
date: 2019-09-30
weight: 3
description: >
  How to add translations
---

All translations are stored in the top-level `translations` directory.

### Adding Translations To an Existing Language
* Run `make extract` to make sure all strings are up to date
* Add translated strings to the appropriate json files in the 'translations'
  directory.

### Adding a New Language
* Add a new json file with the locale code of the language you want to add
  translations for, e.g. en for English.
* Run `make extract` to populate that file with the strings to translate in json
  form.
* Add translations to as many strings as you'd like.
