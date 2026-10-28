# Data provenance and limitations

The classifier was designed with reference to the public `UpdatedResumeDataSet.csv` dataset, available at:
https://github.com/aniketnandi/automated-resume-screening/blob/main/data/UpdatedResumeDataSet.csv

The public dataset is described by its hosting repository as 962 resume records across 25 professional categories, with only 166 unique resume texts because many records are duplicates. The bundled `testdata/resume_benchmark.csv` is **not** a literal excerpt from the full dataset: it is a compact 25-row smoke-test dataset of synthetic skill summaries manually composed to cover the public dataset labels/categories. No real candidate profile is included in this compact test file. It is useful to test the program's CSV pipeline and observable rule behavior, but it is far too small and curated to claim generalizable real-world accuracy. The original CSV is not bundled because it contains personal resume details; a download script is provided to run the full category-routing evaluation locally.

For a more meaningful evaluation, download the original public CSV from the URL below into this directory (filename `UpdatedResumeDataSet.csv`) and run:

```bash
go run . -mode eval -dataset data/UpdatedResumeDataSet.csv
```

The evaluation removes duplicate resume texts before scoring, to avoid repeatedly counting duplicate copies of the same CV. The output evaluates **role-category routing**, not whether an applicant should be hired. The final hire/no-hire rule module has no valid ground-truth outcome labels in this dataset and is instead tested with a decision truth table and parameter-sensitivity experiments.

Raw dataset URL:
https://raw.githubusercontent.com/aniketnandi/automated-resume-screening/main/data/UpdatedResumeDataSet.csv
