import { Component, Input, Output, EventEmitter, signal, computed, ViewChild, ElementRef, HostListener, forwardRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

export interface SelectOption {
  label: string;
  value: any;
  sublabel?: string;
  [key: string]: any;
}

@Component({
  selector: 'app-searchable-select',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './searchable-select.html',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => SearchableSelectComponent),
      multi: true
    }
  ]
})
export class SearchableSelectComponent implements ControlValueAccessor {
  @Input() set options(val: SelectOption[]) {
    this._options.set(val || []);
    // Re-evaluate selected option if options change
    if (this._value()) {
      const match = val?.find(o => o.value === this._value());
      this.selectedOption.set(match || null);
    }
  }
  @Input() placeholder = 'Select an option...';
  @Input() searchPlaceholder = 'Search...';
  @Input() clearable = true;
  @Input() disabled = false;

  @Output() selectionChange = new EventEmitter<SelectOption | null>();

  @ViewChild('searchInput') searchInput!: ElementRef<HTMLInputElement>;

  private _options = signal<SelectOption[]>([]);
  isOpen = signal(false);
  searchQuery = signal('');
  selectedOption = signal<SelectOption | null>(null);
  highlightedIndex = signal(0);
  private _value = signal<any>(null);

  // ControlValueAccessor methods
  onChange: any = () => {};
  onTouch: any = () => {};

  filteredOptions = computed(() => {
    const query = this.searchQuery().toLowerCase().trim();
    if (!query) return this._options();
    return this._options().filter(o => 
      (o.label && o.label.toLowerCase().includes(query)) ||
      (o.sublabel && o.sublabel.toLowerCase().includes(query))
    );
  });

  @HostListener('document:click')
  onDocumentClick() {
    this.close();
  }

  toggleOpen() {
    if (this.disabled) return;
    if (this.isOpen()) {
      this.close();
    } else {
      this.open();
    }
  }

  open() {
    this.isOpen.set(true);
    this.searchQuery.set('');
    this.highlightedIndex.set(0);
    setTimeout(() => {
      this.searchInput?.nativeElement.focus();
    });
  }

  close() {
    this.isOpen.set(false);
    this.onTouch();
  }

  onSearch(event: Event) {
    const val = (event.target as HTMLInputElement).value;
    this.searchQuery.set(val);
    this.highlightedIndex.set(0);
  }

  selectOption(option: SelectOption) {
    this.selectedOption.set(option);
    this._value.set(option.value);
    this.onChange(option.value);
    this.selectionChange.emit(option);
    this.close();
  }

  clearSelection(event: Event) {
    event.stopPropagation();
    this.selectedOption.set(null);
    this._value.set(null);
    this.onChange(null);
    this.selectionChange.emit(null);
  }

  onKeyDown(event: Event) {
    event.preventDefault();
    if (this.highlightedIndex() < this.filteredOptions().length - 1) {
      this.highlightedIndex.update(i => i + 1);
      this.scrollToHighlight();
    }
  }

  onKeyUp(event: Event) {
    event.preventDefault();
    if (this.highlightedIndex() > 0) {
      this.highlightedIndex.update(i => i - 1);
      this.scrollToHighlight();
    }
  }

  onEnter(event: Event) {
    event.preventDefault();
    const opts = this.filteredOptions();
    if (opts.length > 0 && opts[this.highlightedIndex()]) {
      this.selectOption(opts[this.highlightedIndex()]);
    }
  }

  private scrollToHighlight() {
    // Basic implementation; could add actual DOM scrolling logic if needed
  }

  writeValue(value: any): void {
    this._value.set(value);
    const match = this._options().find(o => o.value === value);
    this.selectedOption.set(match || null);
  }
  registerOnChange(fn: any): void { this.onChange = fn; }
  registerOnTouched(fn: any): void { this.onTouch = fn; }
  setDisabledState?(isDisabled: boolean): void { this.disabled = isDisabled; }
}
