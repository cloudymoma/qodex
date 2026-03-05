import { useUIState } from '@/contexts/UIStateContext';
import { useGraphData } from '@/contexts/GraphDataContext';
import { FileTree } from './FileTree';
import { FolderOpen } from 'lucide-react';

export function LeftSidebar() {
  const { treeData, repoName, setCodeViewPath } = useUIState();
  const { focusNode, focusedNodeIds } = useGraphData();

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-dark-border">
        <FolderOpen size={16} className="text-accent-primary" />
        <span className="text-sm font-medium text-dark-text truncate">
          {repoName || 'File Explorer'}
        </span>
      </div>

      {/* Tree */}
      <div className="flex-1 overflow-auto px-1 py-1">
        {treeData.length > 0 ? (
          <FileTree
            data={treeData}
            onFileClick={(path) => focusNode(path)}
            onFileDblClick={(path) => setCodeViewPath(path)}
            focusedPaths={focusedNodeIds}
          />
        ) : (
          <div className="flex items-center justify-center h-full text-dark-text-secondary text-sm">
            Load a repository to explore
          </div>
        )}
      </div>
    </div>
  );
}
