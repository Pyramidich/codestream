const extToLanguage: Record<string, string> = {
  '.js': 'javascript',
  '.jsx': 'javascript',
  '.ts': 'typescript',
  '.tsx': 'typescript',
  '.html': 'html',
  '.css': 'css',
  '.json': 'json',
  '.py': 'python',
  '.go': 'go',
  '.rs': 'rust',
  '.md': 'markdown',
}

export function detectLanguage(filename: string): string {
  const lower = filename.toLowerCase()
  for (const [ext, lang] of Object.entries(extToLanguage)) {
    if (lower.endsWith(ext)) return lang
  }
  return 'plaintext'
}

export const LANGUAGE_OPTIONS = [
  { value: 'plaintext', label: 'Plain Text' },
  { value: 'javascript', label: 'JavaScript' },
  { value: 'typescript', label: 'TypeScript' },
  { value: 'html', label: 'HTML' },
  { value: 'css', label: 'CSS' },
  { value: 'json', label: 'JSON' },
  { value: 'python', label: 'Python' },
  { value: 'go', label: 'Go' },
  { value: 'rust', label: 'Rust' },
  { value: 'markdown', label: 'Markdown' },
]
