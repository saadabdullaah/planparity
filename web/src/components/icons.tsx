import {
  ArchiveIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  CheckCircledIcon,
  ChevronDownIcon,
  CodeIcon,
  CommitIcon,
  Cross1Icon,
  CubeIcon,
  DesktopIcon,
  DownloadIcon,
  EnterFullScreenIcon,
  ExclamationTriangleIcon,
  FileIcon,
  GearIcon,
  HamburgerMenuIcon,
  HomeIcon,
  LayersIcon,
  LockClosedIcon,
  LoopIcon,
  MagicWandIcon,
  MagnifyingGlassIcon,
  MinusIcon,
  MixerHorizontalIcon,
  PlusIcon,
  ReloadIcon,
  RowsIcon,
  Share2Icon,
  StackIcon,
  StopIcon,
  TokensIcon,
} from '@radix-ui/react-icons'
import type { ElementType, SVGAttributes } from 'react'

type IconProps = SVGAttributes<SVGElement> & { size?: number | string; weight?: string }

function icon(Component: ElementType) {
  return function PlanParityIcon({ size = 16, weight, ...props }: IconProps) {
    void weight
    return <Component width={size} height={size} {...props} />
  }
}

export const Archive = icon(ArchiveIcon)
export const ArrowLeft = icon(ArrowLeftIcon)
export const ArrowRight = icon(ArrowRightIcon)
export const ArrowsLeftRight = icon(LoopIcon)
export const BracketsCurly = icon(CodeIcon)
export const CaretDown = icon(ChevronDownIcon)
export const CheckCircle = icon(CheckCircledIcon)
export const CircleNotch = icon(ReloadIcon)
export const CornersOut = icon(EnterFullScreenIcon)
export const Cpu = icon(DesktopIcon)
export const Database = icon(StackIcon)
export const DownloadSimple = icon(DownloadIcon)
export const Flask = icon(MagicWandIcon)
export const FolderOpen = icon(FileIcon)
export const FunnelSimple = icon(MixerHorizontalIcon)
export const Gauge = icon(MixerHorizontalIcon)
export const Graph = icon(Share2Icon)
export const HardDrives = icon(ArchiveIcon)
export const House = icon(HomeIcon)
export const List = icon(HamburgerMenuIcon)
export const ListChecks = icon(RowsIcon)
export const LockKey = icon(LockClosedIcon)
export const LockSimple = icon(LockClosedIcon)
export const MagnifyingGlass = icon(MagnifyingGlassIcon)
export const Minus = icon(MinusIcon)
export const Package = icon(CubeIcon)
export const Path = icon(CommitIcon)
export const Plus = icon(PlusIcon)
export const Scales = icon(TokensIcon)
export const ShieldCheck = icon(CheckCircledIcon)
export const ShieldWarning = icon(ExclamationTriangleIcon)
export const StopCircle = icon(StopIcon)
export const TreeStructure = icon(LayersIcon)
export const WarningCircle = icon(ExclamationTriangleIcon)
export const Wrench = icon(GearIcon)
export const X = icon(Cross1Icon)
