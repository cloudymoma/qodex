import { useState } from 'react';
import type { TreeNode } from '@/types';
import { ChevronRight, ChevronDown, File, Folder } from 'lucide-react';
import clsx from 'clsx';

interface FileTreeProps {
  data: TreeNode[];
  onFileClick: (path: string) => void;
  onFileDblClick: (path: string) => void;
  focusedPaths: ReadonlySet<string>;
  depth?: number;
}

export function FileTree({ data, onFileClick, onFileDblClick, focusedPaths, depth = 0 }: FileTreeProps) {
  return (
    <ul className="list-none m-0 p-0">
      {data.map((node) => (
        <TreeItem
          key={node.path}
          node={node}
          onFileClick={onFileClick}
          onFileDblClick={onFileDblClick}
          focusedPaths={focusedPaths}
          depth={depth}
        />
      ))}
    </ul>
  );
}

function TreeItem({
  node,
  onFileClick,
  onFileDblClick,
  focusedPaths,
  depth,
}: {
  node: TreeNode;
  onFileClick: (path: string) => void;
  onFileDblClick: (path: string) => void;
  focusedPaths: ReadonlySet<string>;
  depth: number;
}) {
  const [expanded, setExpanded] = useState(depth < 1);
  const isDir = node.type === 'directory';
  const isFocused = !isDir && focusedPaths.has(node.path);

  const handleClick = () => {
    if (isDir) {
      setExpanded(!expanded);
    } else {
      onFileClick(node.path);
    }
  };

  const handleDoubleClick = () => {
    if (!isDir) {
      onFileDblClick(node.path);
    }
  };

  return (
    <li>
      <button
        onClick={handleClick}
        onDoubleClick={handleDoubleClick}
        className={clsx(
          'flex items-center gap-1 w-full px-2 py-0.5 text-sm rounded',
          'transition-colors text-left',
          isFocused
            ? 'bg-accent-primary/15 text-accent-primary'
            : 'hover:bg-dark-bg-tertiary text-dark-text-secondary hover:text-dark-text',
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
      >
        {isDir ? (
          expanded ? (
            <ChevronDown size={14} />
          ) : (
            <ChevronRight size={14} />
          )
        ) : (
          <span className="w-3.5" />
        )}
        {isDir ? (
          <Folder size={14} className="text-accent-warning shrink-0" />
        ) : (
          <File size={14} className={clsx('shrink-0', isFocused ? 'text-accent-primary' : 'text-dark-text-secondary')} />
        )}
        <span className="truncate">{node.name}</span>
      </button>

      {isDir && expanded && node.children && (
        <FileTree
          data={node.children}
          onFileClick={onFileClick}
          onFileDblClick={onFileDblClick}
          focusedPaths={focusedPaths}
          depth={depth + 1}
        />
      )}
    </li>
  );
}
