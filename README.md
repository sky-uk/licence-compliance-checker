# Licence compliance checker

Command line application to validate that project dependencies comply with licence restrictions dictated by a company.
This works by identifying project licences using the reliable [go-license-detector](https://github.com/src-d/go-license-detector),
then validating that the most pertinent licences (those with highest confidence level) do not match any of the restricted licences specified as arguments. 
The list of restricted licences are passed as arguments to allow projects to have different restriction policies.
Optional filters can be used to ignore some projects and/or override the licence detected for a project.   

## Usage

```
licence-compliance-checker -r LGPL -r GPL -r AGPL -o vendor/github.com/spf13/cobra=MIT vendor/github.com/spf13/cobra vendor/golang.org/x/crypto
```

Input argument | Meaning 
---------|---------
--restricted-licence (-r) | The name of the licence to restrict. Repeat this flag to specify multiple values.
--ignore-project (-i) | Project which licence will not be checked for compliance. Repeat this flag to specify multiple values.
--override-licence (-o) | Can be used to override the licence detected for a project - e.g. github.com/spf13/cobra=MIT. Repeat this flag to specify multiple values.

Output argument | Meaning 
---------|---------
--log-level (-L) | should be one of: (none), debug, info, warn, error, fatal, panic. default (none)
--show-compliance-all (-A) | to show compliance checks as JSON regardless of outcome. default (false)
--show-compliance-errors (-E) | to show compliance checks as JSON only in case of errors. default (false)

Example of JSON output
```json
{
  "compliant": [
    {
      "project": "testdata/BSD3",
      "matches": [
        {
          "license": "BSD-3-Clause",
          "confidence": 0.9953052
        },
        {
          "license": "BSD-4-Clause",
          "confidence": 0.84976524
        },
        {
          "license": "BSD-3-Clause-No-Nuclear-License-2014",
          "confidence": 0.8356807
        },
        {
          "license": "BSD-Source-Code",
          "confidence": 0.8333333
        }
      ]
    }
  ],
  "restricted": [
    {
      "project": "testdata/MIT",
      "matches": [
        {
          "license": "MIT",
          "confidence": 0.9814815
        },
        {
          "license": "JSON",
          "confidence": 0.9259259
        },
        {
          "license": "MIT-feh",
          "confidence": 0.8447205
        },
        {
          "license": "Xnet",
          "confidence": 0.80864197
        }
      ]
    }
  ],
  "unidentifiable": null,
  "ignored": null
}

```


## Development

To setup the environment with the required dependencies:
```
./make setup
```
To build and run all tests:

```
./make install check
```

## Releasing

Tag the commit in master and push it to release it. Only maintainers can do this.
