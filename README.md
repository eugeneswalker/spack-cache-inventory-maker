# Spack Cache Inventory Maker


## Instructions


1. Edit s3.secrets.tpl, save as s3.secrets


2. Download the specfiles
```
$> cd 0-download
$> ./build.sh

# Adjust parallelism as appropriate in run.sh
$> ./run.sh

$> cd ..
```


3. Parse the downloaded specfiles
```
$> cd 1-parse
$> ./build.sh

# Adjust parallelism as appropriate in run-0-parse.sh
$> ./run-0-parse.sh
$> ./run-1-add-modtimes.sh

$> cd ..
```


4. Generate the static HTML
```
$> cd 2-generate-index
$> ./build.sh
$> ./run.sh
$> cd ..
```

5. Static HTML files will be found in
```
data/html
   /index.html
   /packages
      /ADIAK.html
      /ADIOS2.html
      /ADIOS.html
      /ADLBX.html
      ...
```
