import * as React from "react"
import { cn } from "@/lib/utils"

interface SliderProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'onChange'> {
  value: number[]
  onValueChange: (value: number[]) => void
  min?: number
  max?: number
  step?: number
  minStepsBetweenThumbs?: number
  disabled?: boolean
}

const Slider = React.forwardRef<HTMLDivElement, SliderProps>(
  ({ 
    className, 
    value, 
    onValueChange, 
    min = 0, 
    max = 100, 
    step = 1,
    minStepsBetweenThumbs = 0,
    disabled = false,
    ...props 
  }, ref) => {
    const trackRef = React.useRef<HTMLDivElement>(null)
    const [dragging, setDragging] = React.useState<number | null>(null)

    const getPercentage = (val: number) => {
      return ((val - min) / (max - min)) * 100
    }

    const getValueFromPosition = (clientX: number) => {
      if (!trackRef.current) return value[0]
      
      const rect = trackRef.current.getBoundingClientRect()
      const percentage = Math.max(0, Math.min(100, ((clientX - rect.left) / rect.width) * 100))
      const rawValue = (percentage / 100) * (max - min) + min
      const steppedValue = Math.round(rawValue / step) * step
      return Math.max(min, Math.min(max, steppedValue))
    }

    const handlePointerDown = (e: React.PointerEvent, thumbIndex: number) => {
      if (disabled) return
      e.preventDefault()
      setDragging(thumbIndex)
      ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
    }

    const handlePointerMove = (e: React.PointerEvent) => {
      if (dragging === null || disabled) return
      
      const newValue = getValueFromPosition(e.clientX)
      const newValues = [...value]
      
      if (value.length === 2) {
        // Range slider
        if (dragging === 0) {
          newValues[0] = Math.min(newValue, value[1] - minStepsBetweenThumbs * step)
        } else {
          newValues[1] = Math.max(newValue, value[0] + minStepsBetweenThumbs * step)
        }
      } else {
        newValues[0] = newValue
      }
      
      onValueChange(newValues)
    }

    const handlePointerUp = (e: React.PointerEvent) => {
      if (dragging !== null) {
        ;(e.target as HTMLElement).releasePointerCapture(e.pointerId)
        setDragging(null)
      }
    }

    const handleTrackClick = (e: React.MouseEvent) => {
      if (disabled) return
      
      const newValue = getValueFromPosition(e.clientX)
      
      if (value.length === 2) {
        // Find closest thumb
        const distToFirst = Math.abs(newValue - value[0])
        const distToSecond = Math.abs(newValue - value[1])
        
        if (distToFirst < distToSecond) {
          onValueChange([Math.min(newValue, value[1] - minStepsBetweenThumbs * step), value[1]])
        } else {
          onValueChange([value[0], Math.max(newValue, value[0] + minStepsBetweenThumbs * step)])
        }
      } else {
        onValueChange([newValue])
      }
    }

    // Calculate filled range
    const rangeStart = value.length === 2 ? getPercentage(value[0]) : 0
    const rangeEnd = value.length === 2 ? getPercentage(value[1]) : getPercentage(value[0])

    return (
      <div
        ref={ref}
        className={cn(
          "relative flex w-full touch-none select-none items-center",
          disabled && "opacity-50",
          className
        )}
        {...props}
      >
        <div
          ref={trackRef}
          className="relative h-2 w-full grow overflow-hidden rounded-full bg-secondary cursor-pointer"
          onClick={handleTrackClick}
        >
          <div
            className="absolute h-full bg-primary"
            style={{
              left: `${rangeStart}%`,
              width: `${rangeEnd - rangeStart}%`,
            }}
          />
        </div>
        {value.map((val, index) => (
          <div
            key={index}
            className={cn(
              "absolute block h-5 w-5 rounded-full border-2 border-primary bg-background ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
              dragging === index && "ring-2 ring-ring ring-offset-2",
              !disabled && "cursor-grab active:cursor-grabbing"
            )}
            style={{
              left: `calc(${getPercentage(val)}% - 10px)`,
            }}
            onPointerDown={(e) => handlePointerDown(e, index)}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerUp}
            tabIndex={disabled ? -1 : 0}
            role="slider"
            aria-valuemin={min}
            aria-valuemax={max}
            aria-valuenow={val}
          />
        ))}
      </div>
    )
  }
)
Slider.displayName = "Slider"

export { Slider }



