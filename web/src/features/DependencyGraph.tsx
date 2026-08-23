import { CornersOut, MagnifyingGlass, Minus, Plus } from '../components/icons'
import cytoscape, { type Core, type ElementDefinition, type EventObject, type NodeSingular } from 'cytoscape'
import { useEffect, useMemo, useRef, useState } from 'react'
import type { ModuleReference, ProbableMove, ResourceInstance } from '../api/types'

type ChangeTone = 'added' | 'removed' | 'changed' | 'moved' | 'unchanged'

export function DependencyGraph({ resources, modules = [], moves = [], tones = {} }: {
  resources: ResourceInstance[]
  modules?: ModuleReference[]
  moves?: ProbableMove[]
  tones?: Record<string, ChangeTone>
}) {
  const container = useRef<HTMLDivElement>(null)
  const graph = useRef<Core | null>(null)
  const [selectedAddress, setSelectedAddress] = useState('')
  const selected = resources.find((resource) => resource.address === selectedAddress)
  const elements = useMemo<ElementDefinition[]>(() => {
    const items: ElementDefinition[] = []
    const resourceAddresses = new Set(resources.map((resource) => resource.address))
    const moduleAddresses = new Set(modules.map((module) => module.address || 'root'))
    moduleAddresses.add('root')
    resources.forEach((resource) => moduleAddresses.add(resource.module_address || 'root'))
    ;[...moduleAddresses].sort().forEach((address) => items.push({ data: { id: 'module:' + address, label: address === 'root' ? 'root module' : address, kind: 'module' } }))
    resources.forEach((resource) => {
      items.push({ data: { id: 'resource:' + resource.address, address: resource.address, label: resource.resource_name, subtitle: resource.resource_type, kind: 'resource', tone: tones[resource.address] ?? 'unchanged' } })
      items.push({ data: { id: 'contains:' + resource.address, source: 'module:' + (resource.module_address || 'root'), target: 'resource:' + resource.address, kind: 'containment' } })
      ;(resource.dependencies ?? []).filter((dependency) => resourceAddresses.has(dependency)).forEach((dependency) => items.push({ data: { id: 'dependency:' + dependency + ':' + resource.address, source: 'resource:' + dependency, target: 'resource:' + resource.address, kind: 'dependency' } }))
    })
    moves.forEach((move) => {
      if (resourceAddresses.has(move.from_address) && resourceAddresses.has(move.to_address)) items.push({ data: { id: 'move:' + move.from_address + ':' + move.to_address, source: 'resource:' + move.from_address, target: 'resource:' + move.to_address, kind: 'move' } })
    })
    return items
  }, [modules, moves, resources, tones])

  useEffect(() => {
    if (!container.current || !elements.length) return
    const instance = cytoscape({
      container: container.current,
      elements,
      minZoom: 0.25,
      maxZoom: 2.5,
      layout: { name: 'breadthfirst', directed: true, padding: 42, spacingFactor: 1.2 },
      style: [
        { selector: 'node', style: { 'background-color': '#f3f4ef', 'border-color': '#cdd2cb', 'border-width': 1, color: '#1d2421', label: 'data(label)', 'font-family': 'Segoe UI, system-ui, sans-serif', 'font-size': 10, 'text-valign': 'bottom', 'text-margin-y': 7, width: 34, height: 34 } },
        { selector: 'node[kind = "module"]', style: { shape: 'round-rectangle', width: 92, height: 36, 'background-color': '#222a26', 'border-color': '#222a26', color: '#f5f6f1', 'text-valign': 'center', 'text-margin-y': 0 } },
        { selector: 'node[tone = "added"]', style: { 'background-color': '#4090e0', 'border-color': '#1d6fbd' } },
        { selector: 'node[tone = "removed"]', style: { 'background-color': '#d96055', 'border-color': '#a94139' } },
        { selector: 'node[tone = "changed"]', style: { 'background-color': '#e9a63a', 'border-color': '#b57b1e' } },
        { selector: 'node[tone = "moved"]', style: { 'background-color': '#8567be', 'border-color': '#624991' } },
        { selector: 'node:selected', style: { 'border-width': 3, 'border-color': '#111713', 'overlay-opacity': 0 } },
        { selector: 'edge', style: { width: 1.2, 'line-color': '#c7cdc7', 'target-arrow-color': '#c7cdc7', 'target-arrow-shape': 'triangle', 'curve-style': 'bezier', opacity: 0.72 } },
        { selector: 'edge[kind = "containment"]', style: { 'line-style': 'dashed', 'target-arrow-shape': 'none', opacity: 0.45 } },
        { selector: 'edge[kind = "move"]', style: { width: 2.5, 'line-color': '#8567be', 'target-arrow-color': '#8567be', 'line-style': 'dashed', opacity: 1 } },
      ],
    })
    graph.current = instance
    instance.on('tap', 'node[kind = "resource"]', (event: EventObject) => {
      const node = event.target as NodeSingular
      setSelectedAddress(String(node.data('address')))
    })
    const observer = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(() => instance.resize())
    if (observer && container.current) observer.observe(container.current)
    return () => { observer?.disconnect(); instance.destroy(); graph.current = null }
  }, [elements])

  if (!resources.length) return <div className="emptyState"><MagnifyingGlass size={27} /><strong>No graphable resources</strong><p>This result does not contain resource topology.</p></div>

  return <div className="graphFrame">
    <div className="graphCanvas" ref={container} aria-label="Resource dependency graph" />
    <div className="graphControls" aria-label="Graph controls">
      <button type="button" onClick={() => graph.current?.zoom({ level: Math.min(2.5, (graph.current?.zoom() ?? 1) + 0.2), renderedPosition: { x: 300, y: 220 } })} aria-label="Zoom in"><Plus size={16} /></button>
      <button type="button" onClick={() => graph.current?.zoom({ level: Math.max(0.25, (graph.current?.zoom() ?? 1) - 0.2), renderedPosition: { x: 300, y: 220 } })} aria-label="Zoom out"><Minus size={16} /></button>
      <button type="button" onClick={() => graph.current?.fit(undefined, 36)} aria-label="Fit graph"><CornersOut size={16} /></button>
    </div>
    <div className="graphLegend"><span><i className="legendModule" />Module</span><span><i />Resource</span>{moves.length > 0 && <span><i className="legendMove" />Probable move</span>}</div>
    {selected && <aside className="graphInspector"><button type="button" onClick={() => setSelectedAddress('')} aria-label="Close graph detail">Close</button><span>Selected resource</span><strong>{selected.address}</strong><dl><div><dt>Type</dt><dd>{selected.resource_type}</dd></div><div><dt>Provider</dt><dd>{selected.provider_address}</dd></div><div><dt>Dependencies</dt><dd>{selected.dependencies?.length ?? 0}</dd></div><div><dt>Attributes</dt><dd>{selected.structure.keys?.length ?? 0} structural keys</dd></div></dl></aside>}
  </div>
}

export default DependencyGraph
