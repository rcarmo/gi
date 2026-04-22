function readTrimmedString(...values: unknown[]): string | null {
  for (const value of values) {
    if (typeof value === 'string' && value.trim()) {
      return value.trim();
    }
  }
  return null;
}

function stripOuterQuotes(value: string): string {
  if (
    (value.startsWith('"') && value.endsWith('"')) ||
    (value.startsWith("'") && value.endsWith("'"))
  ) {
    return value.slice(1, -1);
  }
  return value;
}

export function extractShellCwdFromCommand(command: unknown): string | null {
  if (typeof command !== 'string' || !command.trim()) return null;
  const match = command.match(/^\s*cd\s+(.+?)(?:\s*(?:&&|;|\n))/s);
  if (!match?.[1]) return null;
  const candidate = stripOuterQuotes(match[1].trim());
  return candidate || null;
}

export function extractToolContextPath(toolName: unknown, args: unknown): string | null {
  const record = args && typeof args === 'object' ? args as Record<string, unknown> : null;
  if (!record) return null;

  const cwd = readTrimmedString(record.cwd, record.working_directory, record.workingDirectory);
  if (cwd) return cwd;

  const explicitRepoContext = readTrimmedString(
    record.project_dir,
    record.projectDir,
    record.repo_path,
    record.repoPath,
  );
  if (explicitRepoContext) return explicitRepoContext;

  const command = readTrimmedString(record.command);
  const commandCwd = extractShellCwdFromCommand(command);
  if (commandCwd) return commandCwd;

  if (Array.isArray(record.commands)) {
    for (const entry of record.commands) {
      const nestedCwd = extractShellCwdFromCommand(entry);
      if (nestedCwd) return nestedCwd;
    }
  }

  return null;
}

export function stripRemotePathFromSshTarget(value: unknown): string | null {
  const target = readTrimmedString(value);
  if (!target) return null;
  const match = target.match(/^(.*?):((?:\/|~).*)$/);
  return (match?.[1] || target).trim() || null;
}

export function extractToolSshTarget(toolName: unknown, args: unknown): string | null {
  const normalizedToolName = typeof toolName === 'string' ? toolName.trim() : '';
  const record = args && typeof args === 'object' ? args as Record<string, unknown> : null;
  if (!record) return null;

  if (normalizedToolName !== 'ssh') {
    return stripRemotePathFromSshTarget(
      readTrimmedString(record.ssh_target, record.sshTarget, record.remote, record.host)
    );
  }

  return stripRemotePathFromSshTarget(
    readTrimmedString(record.ssh_target, record.sshTarget, record.remote, record.host)
  );
}

export function extractHostLabelFromUrl(value: unknown): string | null {
  const raw = readTrimmedString(value);
  if (!raw) return null;
  try {
    return new URL(raw).host || null;
  } catch {
    const withoutProtocol = raw.replace(/^[a-z]+:\/\//i, '');
    const host = withoutProtocol.split('/')[0]?.trim();
    return host || null;
  }
}

export function extractToolProxmoxHost(toolName: unknown, args: unknown): string | null {
  const normalizedToolName = typeof toolName === 'string' ? toolName.trim() : '';
  if (normalizedToolName !== 'proxmox') return null;
  const record = args && typeof args === 'object' ? args as Record<string, unknown> : null;
  if (!record) return null;
  return extractHostLabelFromUrl(readTrimmedString(record.base_url, record.baseUrl, record.url, record.endpoint));
}

export function extractToolPortainerHost(toolName: unknown, args: unknown): string | null {
  const normalizedToolName = typeof toolName === 'string' ? toolName.trim() : '';
  if (normalizedToolName !== 'portainer') return null;
  const record = args && typeof args === 'object' ? args as Record<string, unknown> : null;
  if (!record) return null;
  return extractHostLabelFromUrl(readTrimmedString(record.base_url, record.baseUrl, record.url, record.endpoint));
}

export function extractToolM365Host(toolName: unknown, args: unknown): string | null {
  const normalizedToolName = typeof toolName === 'string' ? toolName.trim() : '';
  if (!normalizedToolName.startsWith('m365_')) return null;
  const record = args && typeof args === 'object' ? args as Record<string, unknown> : null;
  const explicitHost = extractHostLabelFromUrl(readTrimmedString(
    record?.siteUrl,
    record?.site_url,
    record?.webUrl,
    record?.web_url,
    record?.shareUrl,
    record?.share_url,
    record?.base_url,
    record?.baseUrl,
    record?.url,
    record?.endpoint,
  ));
  if (explicitHost) return explicitHost;
  if (normalizedToolName.startsWith('m365_teams_')) return 'teams.microsoft.com';
  return 'graph.microsoft.com';
}
