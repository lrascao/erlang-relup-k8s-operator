import os.path
import tempfile
import shutil
import tarfile
import subprocess
from kubernetes import client, config, watch

def main():
    # Configs can be set in Configuration class directly or using helper utility
    config.load_incluster_config()

    crds = client.CustomObjectsApi()
    w = watch.Watch()
    print("Watching for release upgrade events")

    for event in w.stream(crds.list_cluster_custom_object, 'relup.lrascao.github.io', 'v1alpha1', 'releaseupgrades', resource_version=''):
        print("Event: %s %s" % (event['type'], event['object']))

        # is this a new release upgrade?
        if event['type'] == 'ADDED':
            handle_new_release_upgrade(event['object']['metadata'], event['object']['spec']) 

def handle_new_release_upgrade(metadata, spec):
    print(" deployment: %s" % (spec['deployment']['name']))
    print(" relup:")
    print("     name: %s" % (spec['relup']['name']))
    print("     image: %s" % (spec['relup']['image']))
    print("     tarball: %s" % (spec['relup']['tarball']))
    print("     source version: %s" % (spec['relup']['sourceVersion']))
    print("     target version: %s" % (spec['relup']['targetVersion']))
    print(" volume:")
    print("     hostPath: %s" % (spec['volume']['hostPath']))

    print("relup %s(%s) from %s to %s has been requested" % (metadata['name'], metadata['uid'], spec['relup']['sourceVersion'], spec['relup']['targetVersion']))

    # create a dir to hold the tarball
    tempdir = tempfile.mkdtemp()
    # copy the tarball over from it's configured source location
    tarball = os.path.basename(spec['relup']['tarball'])
    shutil.copyfile(os.path.join(os.environ['UPGRADE_SOURCE_DIR'], tarball), os.path.join(tempdir, tarball))
    print("%s copied over to %s" % (os.path.join(os.environ['UPGRADE_SOURCE_DIR'], tarball), tempdir))

    # move into the temp dir and untar the release upgrade
    os.chdir(tempdir)
    tar = tarfile.open(tarball)
    tar.extractall()
    tar.close()
    print("%s has been untarred" % (tarball))

    # find the process id of the currently running release
    release_name = os.environ['RELEASE_NAME']
    proc = subprocess.Popen(['bin/{0}'.format(release_name), 'pid'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = proc.communicate()
    if (proc.returncode != 0):
        print('unable to obtain %s pid (code(%s), error: %s)'  % (release_name, str(proc.returncode), error.decode('ascii')))
        return
    pid = output.decode('ascii')
    print("%s pid: %s" % (release_name, pid))

    # now that we have the pid of the Erlang VM running the application on hand we can copy over the release upgrade to
    # it's filesystem namespace
    shutil.copyfile(os.path.join(os.environ['UPGRADE_SOURCE_DIR'], tarball),
                    os.path.join('/proc', pid.strip(), 'root', os.path.relpath(os.environ['RELEASE_ROOT_DIR'], '/'), 'releases', tarball))
    print("%s copied over to %s application's fs process space at %s" % (os.path.join(os.environ['UPGRADE_SOURCE_DIR'], tarball), release_name, os.path.join('/proc', pid.strip(), 'root', os.path.relpath(os.environ['RELEASE_ROOT_DIR'], '/'), 'releases')))

    # and finally ask the application to perform the release upgrade
    cmd = "os:cmd(\"bin/{0} upgrade {1}\").".format(release_name, spec['relup']['targetVersion'])
    proc = subprocess.Popen(['bin/{0}'.format(release_name), 'eval', cmd], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = proc.communicate()
    if (proc.returncode != 0):
        print('unable to eval %s (code(%s), error: %s)'  % (cmd, str(proc.returncode), error.decode('ascii')))
        return
    output = output.decode('ascii')
    print("%s versions: %s" % (release_name, output))

    # delete the temp dir
    # shutil.rmtree(tempdir)

if __name__ == "__main__":
    main()

# {
#     'apiVersion': 'relup.lrascao.github.io/v1alpha1',
#     'kind': 'ReleaseUpgrade',
#     'metadata': {
#         'annotations': {
#              'kubectl.kubernetes.io/last-applied-configuration': '{"apiVersion":"relup.lrascao.github.io/v1alpha1","kind":"ReleaseUpgrade","metadata":{"annotations":{},"name":"relup-0-1-13-0-1-14","namespace":"default"},"spec":{"deployment":{"name":"simple-web-service"},"relup":{"image":"simple-web-server-relup:0.1.13-feature-docker_relup_image","name":"relup-0-1-13-0-1-14-img","tarball":"/srv/upgrade/simple_web_server-0.1.14.tar.gz"},"volume":{"hostPath":"/tmp/simple-web-server-upgrades"}}}\n'
#         },
#         'creationTimestamp': '2021-01-07T23:25:19Z',
#         'generation': 1, 
#         'managedFields': [{
#             'apiVersion': 'relup.lrascao.github.io/v1alpha1',
#             'fieldsType': 'FieldsV1',
#             'fieldsV1': {
#                 'f:metadata': {
#                     'f:annotations': {
#                         '.': {},
#                         'f:kubectl.kubernetes.io/last-applied-configuration': {}
#                     }
#                 },
#                 'f:spec': {
#                     '.': {},
#                     'f:deployment': {
#                         '.': {},
#                         'f:name': {}
#                     },
#                     'f:relup': {
#                         '.': {},
#                         'f:image': {},
#                         'f:name': {},
#                         'f:tarball': {}
#                     },
#                     'f:volume': {
#                         '.': {},
#                         'f:hostPath': {}
#                     }
#                 }
#             },
#             'manager': 'kubectl-client-side-apply',
#             'operation': 'Update',
#             'time': '2021-01-07T23:25:19Z'
#         }],
#         'name': 'relup-0-1-13-0-1-14',
#         'namespace': 'default',
#         'resourceVersion': '1393974',
#         'selfLink': '/apis/relup.lrascao.github.io/v1alpha1/namespaces/default/releaseupgrades/relup-0-1-13-0-1-14',
#         'uid': 'bccda1ba-f797-46a1-a6dd-2b83b5fbcbcf'
#     },
#     'spec': {
#         'deployment': {
#             'name': 'simple-web-service'
#         },
#         'relup': {
#             'image': 'simple-web-server-relup:0.1.13-feature-docker_relup_image',
#             'name': 'relup-0-1-13-0-1-14-img',
#             'tarball': '/srv/upgrade/simple_web_server-0.1.14.tar.gz'
#         },
#         'volume': {
#             'hostPath': '/tmp/simple-web-server-upgrades'
#         }
#     }
# }
