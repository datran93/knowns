import { useRef, useState, useCallback } from "react";
import { X, Trash2 } from "lucide-react";
import { type Annotation } from "../../../hooks/useAnnotations";

interface AnnotationBubbleProps {
	annotation: Annotation;
	isEditing: boolean;
	onTextChange: (text: string) => void;
	onClose: () => void;
	onDelete: () => void;
	onEdit: () => void;
	onSave: () => void;
}

export function AnnotationBubble({
	annotation,
	isEditing,
	onTextChange,
	onClose,
	onDelete,
	onEdit,
	onSave,
}: AnnotationBubbleProps) {
	const [position, setPosition] = useState<{ x: number; y: number }>(() => {
		// Load saved position or default to near the highlight
		const saved = localStorage.getItem(`annotation-pos-${annotation.id}`);
		if (saved) {
			try {
				return JSON.parse(saved);
			} catch {
				// fall through
			}
		}
		return { x: window.innerWidth - 320, y: 120 };
	});
	const [isDragging, setIsDragging] = useState(false);
	const dragStartRef = useRef<{ x: number; y: number; posX: number; posY: number } | null>(null);
	const textareaRef = useRef<HTMLTextAreaElement>(null);

	const handleMouseDown = useCallback((e: React.MouseEvent) => {
		if ((e.target as HTMLElement).closest("button, textarea, input")) return;
		e.preventDefault();
		setIsDragging(true);
		dragStartRef.current = { x: e.clientX, y: e.clientY, posX: position.x, posY: position.y };
	}, [position]);

	const handleMouseMove = useCallback((e: MouseEvent) => {
		if (!isDragging || !dragStartRef.current) return;
		const dx = e.clientX - dragStartRef.current.x;
		const dy = e.clientY - dragStartRef.current.y;
		setPosition((prev) => ({
			x: Math.max(0, Math.min(window.innerWidth - 300, prev.x + dx)),
			y: Math.max(0, Math.min(window.innerHeight - 200, prev.y + dy)),
		}));
		dragStartRef.current.x = e.clientX;
		dragStartRef.current.y = e.clientY;
	}, [isDragging]);

	const handleMouseUp = useCallback(() => {
		if (!isDragging) return;
		setIsDragging(false);
		dragStartRef.current = null;
		// Save position
		localStorage.setItem(`annotation-pos-${annotation.id}`, JSON.stringify(position));
	}, [isDragging, position]);

	// Attach global mouse events when dragging
	if (isDragging) {
		document.addEventListener("mousemove", handleMouseMove);
		document.addEventListener("mouseup", handleMouseUp);
	}

	const handleSaveAndClose = () => {
		onSave();
		onClose();
	};

	return (
		<div
			className="annotation-bubble fixed z-[9999] bg-background border border-border rounded-xl shadow-2xl w-80 flex flex-col"
			style={{
				left: `${position.x}px`,
				top: `${position.y}px`,
				pointerEvents: "auto",
			}}
			data-annotation-bubble=""
		>
			{/* Header / drag handle */}
			<div
				className={`flex items-center gap-2 px-3 py-2 border-b border-border cursor-grab select-none rounded-t-xl ${isDragging ? "cursor-grabbing bg-muted/50" : "bg-muted/30"}`}
				onMouseDown={handleMouseDown}
			>
				<span className="flex-1 text-xs font-medium text-muted-foreground truncate">
					Annotation
				</span>
				<button
					onClick={onClose}
					className="p-1 hover:bg-accent rounded transition-colors"
					title="Close"
				>
					<X className="w-3.5 h-3.5 text-muted-foreground" />
				</button>
			</div>

			{/* Color indicator */}
			<div
				className="h-1 rounded-t-none"
				style={{ backgroundColor: annotation.color }}
			/>

			{/* Content */}
			<div className="p-3 flex flex-col gap-2">
				{/* Quoted text */}
				<div className="text-xs italic text-muted-foreground bg-muted/30 rounded px-2 py-1 border-l-2" style={{ borderColor: annotation.color }}>
					"{annotation.quote.length > 80 ? annotation.quote.slice(0, 80) + "..." : annotation.quote}"
				</div>

				{isEditing ? (
					<>
						<textarea
							ref={textareaRef}
							value={annotation.text}
							onChange={(e) => onTextChange(e.target.value)}
							className="w-full min-h-[80px] text-sm bg-muted/50 border border-border rounded-lg px-3 py-2 resize-none focus:outline-none focus:ring-2 focus:ring-primary/30"
							placeholder="Add your annotation..."
							autoFocus
						/>
						<div className="flex items-center gap-2">
							<button
								onClick={handleSaveAndClose}
								className="flex-1 text-xs bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg py-1.5 font-medium transition-colors"
							>
								Save ⌘↵
							</button>
							<button
								onClick={onDelete}
								className="p-1.5 hover:bg-destructive/10 text-destructive rounded-lg transition-colors"
								title="Delete annotation"
							>
								<Trash2 className="w-3.5 h-3.5" />
							</button>
						</div>
					</>
				) : (
					<>
						<p className="text-sm text-foreground leading-relaxed">
							{annotation.text || <span className="italic text-muted-foreground">No annotation text</span>}
						</p>
						<div className="flex items-center gap-2">
							<button
								onClick={onEdit}
								className="flex-1 text-xs text-muted-foreground hover:text-foreground bg-muted/50 hover:bg-muted rounded-lg py-1.5 font-medium transition-colors border border-border/50"
							>
								Edit
							</button>
							<button
								onClick={onDelete}
								className="p-1.5 hover:bg-destructive/10 text-destructive rounded-lg transition-colors"
								title="Delete annotation"
							>
								<Trash2 className="w-3.5 h-3.5" />
							</button>
						</div>
					</>
				)}
			</div>
		</div>
	);
}
