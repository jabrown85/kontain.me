## What is this?

An experimental GCB API-compatible buildpack service running on Cloud Run. A GCB
user can use any recently released `gcloud` along with this service to upload
and build buildpack-compatible source and produce a container image.

Any requested `steps` are ignored and a buildpack build is executed on the
source instead. The request must specify exactly one image to build in `images`,
and must specify a `storageSource`.

This is **an experiment** and should absolutely not be used for anything serious.

## How do I use it?

First, get into a local directory containing buildpack-detectable source:

```
$ git clone git@github.com:buildpack/sample-java-app.git
$ cd sample-java-app
```

Then, by using `gcloud` and overriding the address where API requests are sent,
you can create Build requests that execute buildpacks builds:

```
$ CLOUDSDK_API_ENDPOINT_OVERRIDES_CLOUDBUILD=https://api-an3qnndwmq-uc.a.run.app/ gcloud builds submit --tag=gcr.io/my-project/built
Creating temporary tarball archive of 15 file(s) totalling 91.8 KiB before compression.
Some files were not included in the source upload.

Check the gcloud log [/Users/jasonhall/.config/gcloud/logs/2019.05.16/00.35.06.407646.log] to see which files and the contents of the
default gcloudignore file used (see `$ gcloud topic gcloudignore` to learn
more).

Uploading tarball of [.] to [gs://my-project_cloudbuild/source/1557981306.47-9ee5987ef42e4dc988d7dcd4a4dc0bdc.tgz]
Created [https://api-an3qnndwmq-uc.a.run.app/v1/projects/my-project/builds/a33da1cc-8e3c-4579-92cd-d7bab749ba22].

... snip ...
ID                                    CREATE_TIME                DURATION  SOURCE                                                                                IMAGES                   STATUS
a33da1cc-8e3c-4579-92cd-d7bab749ba22  2019-05-16T04:35:07+00:00  1M6S      gs://my-project_cloudbuild/source/1557981306.47-9ee5987ef42e4dc988d7dcd4a4dc0bdc.tgz  gcr.io/my-project/built  SUCCESS
```

You can also get build details:

```
$ CLOUDSDK_API_ENDPOINT_OVERRIDES_CLOUDBUILD=https://api-an3qnndwmq-uc.a.run.app/ gcloud builds describe a33da1cc-8e3c-4579-92cd-d7bab749ba22
createTime: '2019-05-16T04:35:07.426525401Z'
finishTime: '2019-05-16T04:36:13.70673836Z'
id: a33da1cc-8e3c-4579-92cd-d7bab749ba22
images:
- gcr.io/my-project/built
logsBucket: my-project_cloudbuild
projectId: my-project
results:
  images:
  - digest: sha256:de35ebf2e6e39bc7e2047bc261095435dd6b710ff09af38edcb059e640e8c35e
    name: gcr.io/my-project/built
source:
  storageSource:
    bucket: my-project_cloudbuild
    generation: '1557981307124794'
    object: source/1557981306.47-9ee5987ef42e4dc988d7dcd4a4dc0bdc.tgz
startTime: '2019-05-16T04:35:07.426525401Z'
status: SUCCESS
statusDetail: ''
```

## Known differences

-  Builds are performed entirely in the context of the `projects.builds.create`
   request, not by polling a long-running operation.  The `--async` flag has no
effect.
- Build operations (source pulls and image pushes) are authorized using the
  end-user credentials, not the project's builder service account.
- Build logs are written to Cloud Storage in one shot, at the end of the build.
  When the build request completes, `gcloud` will show all build logs at once
without streaming.
- Build logs are written to the source upload bucket, and not a separate logs
  bucket as is the default in GCB.
- Builds cannot be cancelled. The client doesn't know the build ID until it's
  completed.

## Not yet implemented

- [ ] `timing` is not collected or reported.
- [ ] `timeout` is not configurable. If Cloud Run request times out, client
  gets a 502.
- [ ] `sourceProvenance` is not yet collected or reported, or uploaded to a
  Grafeas instance.
- [ ] `projects.builds.list` is not yet implemented.
- [ ] `operations.get` and `operations.list` are not yet implemented.
