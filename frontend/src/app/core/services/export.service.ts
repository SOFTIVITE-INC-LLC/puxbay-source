import { Injectable, inject } from '@angular/core';
import { ApiService } from './api.service';
import { ToastService } from './toast';

export interface ExportColumn {
  header: string;
  key?: string;
  type?: 'text' | 'currency' | 'number' | 'date' | 'percent';
}

export interface SummaryCard {
  label: string;
  value: string | number;
  highlight?: boolean;
}

export interface ExportOptions {
  filename: string;
  title: string;
  subtitle?: string;
  companyName?: string;
  dateRange?: string;
  headers: string[];
  rows: any[][];
  summaryCards?: SummaryCard[];
  orientation?: 'portrait' | 'landscape';
}

@Injectable({
  providedIn: 'root'
})
export class ExportService {
  private api = inject(ApiService);
  private toast = inject(ToastService);

  /**
   * Export product catalog CSV via backend API
   */
  exportProducts(branchId?: string): void {
    const params: any = {};
    if (branchId) params.branch_id = branchId;
    this.api.get<Blob>('/export/products', { params, responseType: 'blob' as any }).subscribe({
      next: (blob) => {
        this.triggerDownload(blob, `products_export_${this.getDateStamp()}.csv`);
        this.toast.showSuccess('Products catalog exported!');
      },
      error: () => this.toast.showError('Failed to export products CSV')
    });
  }

  /**
   * Export orders CSV via backend API
   */
  exportOrders(branchId?: string, startDate?: string, endDate?: string): void {
    const params: any = {};
    if (branchId) params.branch_id = branchId;
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;
    this.api.get<Blob>('/export/orders', { params, responseType: 'blob' as any }).subscribe({
      next: (blob) => {
        this.triggerDownload(blob, `orders_export_${this.getDateStamp()}.csv`);
        this.toast.showSuccess('Orders exported!');
      },
      error: () => this.toast.showError('Failed to export orders CSV')
    });
  }

  /**
   * Export inventory CSV via backend API
   */
  exportInventory(branchId?: string): void {
    const params: any = {};
    if (branchId) params.branch_id = branchId;
    this.api.get<Blob>('/export/inventory', { params, responseType: 'blob' as any }).subscribe({
      next: (blob) => {
        this.triggerDownload(blob, `inventory_export_${this.getDateStamp()}.csv`);
        this.toast.showSuccess('Inventory exported!');
      },
      error: () => this.toast.showError('Failed to export inventory CSV')
    });
  }

  /**
   * Export dataset to Microsoft Excel format (.xls / XML Spreadsheet)
   */
  exportToExcel(options: ExportOptions): void {
    const { filename, title, subtitle, dateRange, headers, rows, summaryCards } = options;
    const safeFilename = this.sanitizeFilename(filename) + '_' + this.getDateStamp() + '.xls';

    let xml = `<?xml version="1.0" encoding="UTF-8"?>
<?mso-application progid="Excel.Sheet"?>
<Workbook xmlns="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:o="urn:schemas-microsoft-com:office:office"
 xmlns:x="urn:schemas-microsoft-com:office:excel"
 xmlns:ss="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:html="http://www.w3.org/TR/REC-html40">
 <DocumentProperties xmlns="urn:schemas-microsoft-com:office:office">
  <Title>${this.escapeXml(title)}</Title>
  <Created>${new Date().toISOString()}</Created>
 </DocumentProperties>
 <Styles>
  <Style ss:ID="Default" ss:Name="Normal">
   <Alignment ss:Vertical="Center"/>
   <Font ss:FontName="Segoe UI" ss:Size="10" ss:Color="#333333"/>
  </Style>
  <Style ss:ID="TitleStyle">
   <Font ss:FontName="Segoe UI" ss:Size="16" ss:Bold="1" ss:Color="#005b96"/>
   <Alignment ss:Vertical="Center"/>
  </Style>
  <Style ss:ID="SubtitleStyle">
   <Font ss:FontName="Segoe UI" ss:Size="10" ss:Italic="1" ss:Color="#666666"/>
  </Style>
  <Style ss:ID="CardLabel">
   <Font ss:FontName="Segoe UI" ss:Size="9" ss:Bold="1" ss:Color="#777777"/>
   <Interior ss:Color="#F4F6F8" ss:Pattern="Solid"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
   </Borders>
  </Style>
  <Style ss:ID="CardValue">
   <Font ss:FontName="Segoe UI" ss:Size="12" ss:Bold="1" ss:Color="#0f172a"/>
   <Interior ss:Color="#F4F6F8" ss:Pattern="Solid"/>
  </Style>
  <Style ss:ID="HeaderStyle">
   <Font ss:FontName="Segoe UI" ss:Size="10" ss:Bold="1" ss:Color="#FFFFFF"/>
   <Interior ss:Color="#005B96" ss:Pattern="Solid"/>
   <Alignment ss:Horizontal="Center" ss:Vertical="Center"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#03396C"/>
   </Borders>
  </Style>
  <Style ss:ID="RowEven">
   <Interior ss:Color="#FFFFFF" ss:Pattern="Solid"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#F1F5F9"/>
   </Borders>
  </Style>
  <Style ss:ID="RowOdd">
   <Interior ss:Color="#F8FAFC" ss:Pattern="Solid"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#F1F5F9"/>
   </Borders>
  </Style>
  <Style ss:ID="CurrencyCell">
   <NumberFormat ss:Format="GHS #,##0.00"/>
  </Style>
  <Style ss:ID="NumberCell">
   <NumberFormat ss:Format="#,##0"/>
  </Style>
 </Styles>
 <Worksheet ss:Name="Report Data">
  <Table ss:DefaultColumnWidth="120">
`;

    // Title & Meta rows
    xml += `   <Row ss:Height="26">
    <Cell ss:StyleID="TitleStyle"><Data ss:Type="String">${this.escapeXml(title)}</Data></Cell>
   </Row>\n`;

    if (subtitle || dateRange) {
      const sub = [subtitle, dateRange ? `Period: ${dateRange}` : ''].filter(Boolean).join(' | ');
      xml += `   <Row ss:Height="18">
    <Cell ss:StyleID="SubtitleStyle"><Data ss:Type="String">${this.escapeXml(sub)}</Data></Cell>
   </Row>\n`;
    }

    // Summary Cards (if any)
    if (summaryCards && summaryCards.length > 0) {
      xml += `   <Row ss:Height="8"></Row>\n`;
      xml += `   <Row ss:Height="18">\n`;
      for (const card of summaryCards) {
        xml += `    <Cell ss:StyleID="CardLabel"><Data ss:Type="String">${this.escapeXml(card.label)}</Data></Cell>\n`;
      }
      xml += `   </Row>\n`;
      xml += `   <Row ss:Height="22">\n`;
      for (const card of summaryCards) {
        const val = String(card.value ?? '');
        xml += `    <Cell ss:StyleID="CardValue"><Data ss:Type="String">${this.escapeXml(val)}</Data></Cell>\n`;
      }
      xml += `   </Row>\n`;
    }

    xml += `   <Row ss:Height="10"></Row>\n`;

    // Table Headers
    xml += `   <Row ss:Height="24">\n`;
    for (const h of headers) {
      const label = String(h || '').replace(/_/g, ' ').toUpperCase();
      xml += `    <Cell ss:StyleID="HeaderStyle"><Data ss:Type="String">${this.escapeXml(label)}</Data></Cell>\n`;
    }
    xml += `   </Row>\n`;

    // Table Data Rows
    for (let i = 0; i < rows.length; i++) {
      const row = rows[i];
      const styleId = i % 2 === 0 ? 'RowEven' : 'RowOdd';
      xml += `   <Row ss:Height="20">\n`;
      for (const cell of row) {
        const cellVal = cell !== null && cell !== undefined ? cell : '';
        const isNum = typeof cellVal === 'number';
        const dataType = isNum ? 'Number' : 'String';
        xml += `    <Cell ss:StyleID="${styleId}"><Data ss:Type="${dataType}">${this.escapeXml(String(cellVal))}</Data></Cell>\n`;
      }
      xml += `   </Row>\n`;
    }

    xml += `  </Table>
 </Worksheet>
</Workbook>`;

    const blob = new Blob([xml], { type: 'application/vnd.ms-excel;charset=utf-8;' });
    this.triggerDownload(blob, safeFilename);
  }

  /**
   * Export dataset as high-quality printable PDF document via dedicated print renderer
   */
  exportToPdf(options: ExportOptions): void {
    const { title, subtitle, companyName = 'Puxbay Business Report', dateRange, headers, rows, summaryCards, orientation = 'portrait' } = options;

    const printWindow = window.open('', '_blank', 'width=1100,height=850');
    if (!printWindow) {
      alert('Please allow popups to export to PDF');
      return;
    }

    const cardsHtml = (summaryCards && summaryCards.length > 0)
      ? `
      <div class="summary-grid">
        ${summaryCards.map(c => `
          <div class="summary-card ${c.highlight ? 'highlight' : ''}">
            <div class="card-label">${this.escapeHtml(c.label)}</div>
            <div class="card-value">${this.escapeHtml(String(c.value))}</div>
          </div>
        `).join('')}
      </div>
      `
      : '';

    const headersHtml = headers.map(h => {
      const isNum = ['revenue', 'orders', 'discounts', 'tax', 'total', 'amount', 'balance', 'price', 'quantity', 'subtotal', 'items_sold', 'avg_order_value', 'net_sales'].some(k => h.toLowerCase().includes(k));
      return `<th class="${isNum ? 'text-right' : 'text-left'}">${this.escapeHtml(h.replace(/_/g, ' ').toUpperCase())}</th>`;
    }).join('');

    const rowsHtml = rows.map((r, idx) => `
      <tr class="${idx % 2 === 0 ? 'even' : 'odd'}">
        ${r.map((cell, cIdx) => {
          const h = (headers[cIdx] || '').toLowerCase();
          const isNum = typeof cell === 'number' || ['revenue', 'discounts', 'tax', 'total', 'amount', 'balance', 'price', 'subtotal', 'net_sales'].some(k => h.includes(k));
          let displayVal = cell !== null && cell !== undefined ? String(cell) : '–';
          return `<td class="${isNum ? 'text-right font-medium' : 'text-left'}">${this.escapeHtml(displayVal)}</td>`;
        }).join('')}
      </tr>
    `).join('');

    const htmlContent = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>${this.escapeHtml(title)} - PDF Export</title>
  <style>
    @page {
      size: ${orientation === 'landscape' ? 'A4 landscape' : 'A4 portrait'};
      margin: 12mm 12mm 15mm 12mm;
    }
    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    body {
      background: #ffffff;
      color: #0f172a;
      padding: 16px;
      font-size: 11px;
      line-height: 1.4;
    }
    .header-bar {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      border-bottom: 2px solid #005b96;
      padding-bottom: 12px;
      margin-bottom: 16px;
    }
    .brand-title {
      font-size: 18px;
      font-weight: 800;
      color: #005b96;
      letter-spacing: -0.02em;
    }
    .company-name {
      font-size: 11px;
      font-weight: 600;
      color: #64748b;
      margin-top: 2px;
    }
    .report-meta {
      text-align: right;
      font-size: 10px;
      color: #64748b;
    }
    .report-title-section {
      margin-bottom: 16px;
    }
    .main-title {
      font-size: 16px;
      font-weight: 700;
      color: #0f172a;
    }
    .subtitle {
      font-size: 11px;
      color: #64748b;
      margin-top: 3px;
    }
    .summary-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
      gap: 8px;
      margin-bottom: 16px;
    }
    .summary-card {
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 10px 12px;
    }
    .summary-card.highlight {
      background: #f0f9ff;
      border-color: #bae6fd;
    }
    .card-label {
      font-size: 9px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #64748b;
      margin-bottom: 4px;
    }
    .card-value {
      font-size: 14px;
      font-weight: 800;
      color: #0f172a;
    }
    .highlight .card-value {
      color: #005b96;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 16px;
      page-break-inside: auto;
    }
    thead {
      display: table-header-group;
    }
    tr {
      page-break-inside: avoid;
      page-break-after: auto;
    }
    th {
      background: #005b96;
      color: #ffffff;
      font-size: 9.5px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      padding: 8px 10px;
      border: 1px solid #005b96;
    }
    td {
      padding: 7px 10px;
      font-size: 10px;
      border: 1px solid #e2e8f0;
      color: #1e293b;
    }
    tr.even td {
      background: #ffffff;
    }
    tr.odd td {
      background: #f8fafc;
    }
    .text-left { text-align: left; }
    .text-right { text-align: right; }
    .text-center { text-align: center; }
    .font-medium { font-weight: 600; }
    .footer {
      border-top: 1px solid #e2e8f0;
      padding-top: 8px;
      display: flex;
      justify-content: space-between;
      font-size: 9px;
      color: #94a3b8;
      margin-top: 16px;
    }
    @media print {
      body {
        padding: 0;
        -webkit-print-color-adjust: exact;
        print-color-adjust: exact;
      }
      .no-print {
        display: none !important;
      }
    }
    .action-bar {
      position: fixed;
      top: 10px;
      right: 10px;
      background: #ffffff;
      padding: 6px 12px;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
      border: 1px solid #e2e8f0;
      display: flex;
      gap: 8px;
      z-index: 999;
    }
    .btn {
      background: #005b96;
      color: white;
      border: none;
      padding: 6px 12px;
      border-radius: 6px;
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
    }
    .btn-secondary {
      background: #e2e8f0;
      color: #1e293b;
    }
  </style>
</head>
<body>
  <div class="action-bar no-print">
    <button class="btn" onclick="window.print()">Print / Save as PDF</button>
    <button class="btn btn-secondary" onclick="window.close()">Close</button>
  </div>

  <div class="header-bar">
    <div>
      <div class="brand-title">Puxbay</div>
      <div class="company-name">${this.escapeHtml(companyName)}</div>
    </div>
    <div class="report-meta">
      <div><strong>Generated:</strong> ${new Date().toLocaleDateString()} ${new Date().toLocaleTimeString()}</div>
      ${dateRange ? `<div><strong>Period:</strong> ${this.escapeHtml(dateRange)}</div>` : ''}
      <div><strong>Total Records:</strong> ${rows.length}</div>
    </div>
  </div>

  <div class="report-title-section">
    <h1 class="main-title">${this.escapeHtml(title)}</h1>
    ${subtitle ? `<div class="subtitle">${this.escapeHtml(subtitle)}</div>` : ''}
  </div>

  ${cardsHtml}

  <table>
    <thead>
      <tr>${headersHtml}</tr>
    </thead>
    <tbody>
      ${rowsHtml}
    </tbody>
  </table>

  <div class="footer">
    <div>Puxbay Intelligent Commerce Platform</div>
    <div>Confidential & Proprietary</div>
  </div>

  <script>
    // Auto open print dialog after styles apply
    window.onload = function() {
      setTimeout(function() {
        window.print();
      }, 500);
    };
  </script>
</body>
</html>
    `;

    printWindow.document.open();
    printWindow.document.write(htmlContent);
    printWindow.document.close();
  }

  /**
   * Export dataset as CSV file
   */
  exportToCsv(options: { filename: string; headers: string[]; rows: any[][] }): void {
    const { filename, headers, rows } = options;
    const safeFilename = this.sanitizeFilename(filename) + '_' + this.getDateStamp() + '.csv';

    const csvContent = [
      headers.map(h => `"${String(h).replace(/"/g, '""')}"`).join(','),
      ...rows.map(row =>
        row.map(cell => {
          const val = cell !== null && cell !== undefined ? String(cell) : '';
          return `"${val.replace(/"/g, '""')}"`;
        }).join(',')
      )
    ].join('\r\n');

    // Add UTF-8 BOM so Excel opens non-ASCII characters without issues
    const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' });
    this.triggerDownload(blob, safeFilename);
  }

  private triggerDownload(blob: Blob, filename: string): void {
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }

  private sanitizeFilename(name: string): string {
    return (name || 'report').toLowerCase().replace(/[^a-z0-9_-]/g, '_');
  }

  private getDateStamp(): string {
    return new Date().toISOString().slice(0, 10);
  }

  private escapeXml(str: string): string {
    return (str || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&apos;');
  }

  private escapeHtml(str: string): string {
    return (str || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }
}
