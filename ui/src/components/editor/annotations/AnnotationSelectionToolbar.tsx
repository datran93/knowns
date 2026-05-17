import { useEffect, useRef, useState, useCallback } from "react";
import { MessageSquarePlus } from "lucide-react";
import { type Annotation } from "../../../hooks/useAnnotations";

interface AnnotationSelectionToolbarProps {
	containerRef: React.RefObject<HTMLElement | null>;
	onCreateAnnotation: (quote: string, startOffset: number, endOffset: number) => void;
}

interface SelectionState {
	visible: boolean;
	rect: DOMRect | null;
	text: string;
	startOffset: number;
	endOffset: number;
}

const COLORS = [
	"#fef08a",
	"#bbf7d0",
	"#bfdbfe",
	"#fecaca",
	"#e9d5ff",
	"#fed7aa",
	"#fbcfe8",
];

export function AnnotationSelectionToolbar({
	containerRef,
	onCreateAnnotation,
}: AnnotationSelectionToolbarProps) {
	const [selection, setSelection] = useState<SelectionState>({
		visible: false,
		rect: null,
		text: "",
		startOffset: 0,
		endOffset: 0,
	});
	const [selectedColor, setSelectedColor] = useState(COLORS[0]);
	const toolbarRef = useRef<HTMLDivElement>(null);

	const computeSelection = useCallback(() => {
		const sel = window.getSelection();
		if (!sel || sel.isCollapsed || !sel.rangeCount) {
			setSelection((prev) => ({ ...prev, visible: false }));
			return;
		}

		const range = sel.getRangeAt(0);
		const container = containerRef.current;
		if (!container || !container.contains(range.commonAncestorContainer)) {
			setSelection((prev) => ({ ...prev, visible: false }));
			return;
		}

		const text = sel.toString().trim();
		if (!text) {
			setSelection((prev) => ({ ...prev, visible: false }));
			return;
		}

		const rect = range.getBoundingClientRect();
		const startOffset = 0; // simplified — full offset needs container traversal
		const endOffset = text.length;

		setSelection({
			visible: true,
			rect,
			text,
			startOffset,
			endOffset,
		});
	}, [containerRef]);

	useEffect(() => {
		document.addEventListener("mouseup", computeSelection);
		return () => document.removeEventListener("mouseup", computeSelection);
	}, [computeSelection]);

	const handleAnnotate = () => {
		if (!selection.text) return;
		onCreateAnnotation(selection.text, selection.startOffset, selection.endOffset);
		window.getSelection()?.removeAllRanges();
		setSelection((prev) => ({ ...prev, visible: false }));
	};

	if (!selection.visible || !selection.rect) return null;

	const toolbarStyle: React.CSSProperties = {
		position: "fixed",
		left: `${selection.rect.left + selection.rect.width / 2}px`,
		top: `${selection.rect.top - 48}px`,
		transform: "translateX(-50%)",
		zIndex: 10000,
	};

	return (
		<div
			ref={toolbarRef}
			style={toolbarStyle}
			className="flex items-center gap-1 bg-background border border-border rounded-full shadow-xl px-2 py-1.5 animate-fade-in"
		>
			{/* Color picker */}
			<div className="flex items-center gap-0.5 mr-1">
				{COLORS.map((color) => (
					<button
						key={color}
						onClick={() => setSelectedColor(color)}
						className={`w-4 h-4 rounded-full border-2 transition-transform hover:scale-110 ${selectedColor === color ? "scale-125 border-foreground" : "border-transparent"}`}
						style={{ backgroundColor: color }}
						title={`Color: ${color}`}
					/>
				))}
			</div>

			<div className="w-px h-5 bg-border/50" />

			{/* Annotate button */}
			<button
				onClick={handleAnnotate}
				className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground hover:text-foreground px-2 py-1 rounded-full hover:bg-accent transition-colors"
				title="Add annotation"
			>
				<MessageSquarePlus className="w-3.5 h-3.5" />
				<span>Annotate</span>
			</button>
		</div>
	);
}
