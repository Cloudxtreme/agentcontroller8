import imp

ATTRIBUTES = ('descr', 'organization', 'name', 'author',
              'license', 'version', 'category', 'async',
              'queue', 'roles', 'enable', 'period', 'timeout')


def get_info(path):
    """
    Loads and retrieve jumpscript attributes.
    """
    module = imp.load_source(path, path)

    attributes = {}
    for attr in ATTRIBUTES:
        if hasattr(module, attr):
            attributes[attr] = getattr(module, attr)

    return attributes
