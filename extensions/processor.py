from JumpScale import j # NOQA
import time

osis = j.clients.osis.getNamespace('system')


def get_or_create_command(command_guid):
    try:
        return osis.command.get(command_guid)
    except Exception, e:
        if e.eco['exceptionclassname'] == 'KeyError':
            return osis.command.new()
        raise


# Entry point called via the controller to process a received command.
def process_command(command):
    cmd = get_or_create_command(command['id'])

    cmd.guid = command['id']

    for key in ('gid', 'nid', 'cmd', 'roles', 'fanout', 'args', 'data', 'tags'):
        setattr(cmd, key, command[key])

    cmd.starttime = int(time.time() * 1000)
    osis.command.set(cmd)


# Entry point called via the controller to process a receieved result.
def process_result(result):
    cmd = get_or_create_command(result['id'])

    gid = result['gid']
    nid = result['nid']

    job = None
    for _job in cmd.jobs:
        if _job.gid == gid and _job.nid == nid:
            job = _job
            break

    if job is None:
        job = cmd.new_job()

    cmd.guid = result['id']

    for key in ('gid', 'nid', 'data', 'streams', 'level', 'state', 'starttime', 'time', 'tags', 'critical'):
        setattr(job, key, result[key])

    osis.command.set(cmd)
