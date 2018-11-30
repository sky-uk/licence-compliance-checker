# Licence compliance checker

Command line application to validate that project dependencies comply with licence restrictions dictated by a company.
This works by identifying project licences using the reliable [go-license-detector](https://github.com/src-d/go-license-detector),
then validating that the most pertinent licences (those with highest confidence level) do not match any of the restricted licences specified as arguments. 
The list of restricted licences are passed as arguments to allow projects to have different restriction policies.
Optional filters can be used to ignore some projects and/or override the licence detected for a project.  

The checker is intended for use in continuous integration pipelines, to help ensure that projects are complying with
licence restrictions on an ongoing basis.

## Installing

If you already have Go installed, the easiest way of installing is with `go get`:

```
go get github.com/sky-uk/licence-compliance-checker
```

## Usage

```
licence-compliance-checker -r LGPL -r GPL -r AGPL -o vendor/github.com/spf13/cobra=MIT vendor/github.com/spf13/cobra vendor/golang.org/x/crypto
```

See the `licencecheck` target in the [Makefile](Makefile) for an example of how to use with dependencies managed by `go dep`


Exit code | Meaning
----------|--------
0 | No restricted licenses found
1 | Restricted licenses found

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

## Questions or Problems?

- If you have a general question about this project, please create an issue for it. The issue title should be the
  question itself, with any follow-up information in a comment. Add the "question" tag to the issue.
  
- If you think you have found a bug in this project, please create an issue for it. Use the issue title to summarise
  the problems, and supply full steps to reproduce in a comment. Add the "bug" tag to the issue.

## Contributions

See [CONTRIBUTING.md](CONTRIBUTING.md)
