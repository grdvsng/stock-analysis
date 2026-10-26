import re


CAMELCASE_PATTERN = re.compile(r'(.)([A-Z][a-z]+)')

def camel_to_snake(name: str) -> str:
    s1 = CAMELCASE_PATTERN.sub(r'\1_\2', name)

    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()
