from JumpScale import j # NOQA
import time
import json

ERROR_STATES = ('ERROR', 'TIMEOUT')

j.data.models.system.connect2mongo()

def get_or_create_command(command_guid):
    try:
        return j.data.models.system.Command.objects.get(guid=command_guid)
    except j.data.models.system.Command.DoesNotExist as e:
        return j.data.models.system.Command()

# Entry point called via the controller to process a received command.
def process_command(command):
    cmd = get_or_create_command(command['id'])

    cmd.guid = command['id']

    for key in ('gid', 'nid', 'cmd', 'roles', 'fanout', 'args', 'data', 'tags'):
        setattr(cmd, key, command[key])

    cmd.starttime = int(time.time() * 1000)
    cmd.save()


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
        job = j.data.models.system.Job()
        cmd.jobs.append(job)

    cmd.guid = result['id']

    for key in ('gid', 'nid', 'data', 'streams', 'level', 'state', 'starttime', 'time', 'tags', 'critical'):
        setattr(job, key, result[key])

    cmd.save()

    if result['state'] in ERROR_STATES:
        process_error_result(result)


def get_eco(result):
    # critical is the last error message that was received via the process
    # and has 'critical' level. Under jumpscale, this will container the
    # json error object.
    try:
        eco_dict = json.loads(result['critical'])
        return j.errorconditionhandler.getErrorConditionObject(eco_dict)
    except:
        streams = result.get('streams') or ['', '']
        error = result['critical'] or streams[1] or result['data'] or result['state']
        eco = j.errorconditionhandler.getErrorConditionObject(msg=error)
        eco.backtrace = ''
        eco.backtraceDetailed = ''
        return eco


def process_error_result(result):
    gid = result['gid']
    nid = result['nid']

    eco = get_eco(result)

    eco_obj = j.data.models.system.Errorcondition()

    for key in ('pid', 'masterjid', 'epoch', 'appname', 'level', 'type', 'state', 'errormessage',
                'errormessagePub', 'category', 'tags', 'code', 'funcname', 'funcfilename', 'funclinenr',
                'backtrace', 'backtraceDetailed', 'lasttime', 'closetime', 'occurrences'):
        setattr(eco_obj, key, getattr(eco, key))

    eco_obj.gid = gid
    eco_obj.nid = nid
    eco_obj.jid = result['id']

    eco_obj.save()
