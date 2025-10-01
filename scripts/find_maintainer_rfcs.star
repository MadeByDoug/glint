"""Select Maintainer-owned RFC Markdown files for schema validation."""

load_root = "docs/specs"


def _owner_markers():
    markers = ["@maintainers"]
    default_team = consts.get("default_team")
    if default_team:
        marker = default_team if default_team.startswith("@") else "@" + default_team
        if marker not in markers:
            markers.append(marker)
    return markers


def _normalize(path):
    return path.replace("\\", "/")


def _owners_match(contents, markers):
    for marker in markers:
        if ("Owners: " + marker) in contents:
            return True
    return False


def _to_relative(path, normalized_root):
    prefix = normalized_root + "/"
    if normalized_root and path.startswith(prefix):
        return path[len(prefix):]
    return path


def _collect_artifacts():
    markers = _owner_markers()
    normalized_root = _normalize(root).rstrip("/")
    result = []

    for path in walk(root=load_root):
        normalized = _normalize(path)
        if not normalized.endswith(".md"):
            continue
        if ("/" + load_root + "/") not in normalized:
            continue

        contents = read_file(path)
        if not _owners_match(contents, markers):
            continue

        rel = _to_relative(normalized, normalized_root)
        result.append({"path": rel, "type": "file"})

    return result


artifacts = _collect_artifacts()
