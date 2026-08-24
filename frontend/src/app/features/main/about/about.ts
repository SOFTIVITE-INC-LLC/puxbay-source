import { Component, ViewEncapsulation } from '@angular/core';

import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-about',
  standalone: true,
  imports: [RouterModule],
  templateUrl: './about.html',
  encapsulation: ViewEncapsulation.None,
})
export class About {}
