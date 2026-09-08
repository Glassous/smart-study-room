<script setup>
defineProps({ computer: Boolean, power: Boolean })
</script>

<template>
  <span class="plan-desk" aria-hidden="true">
    <span class="desk-legs" />
    <span class="desk-top">
      <template v-if="computer">
        <span class="monitor"><span /></span>
        <span class="keyboard" /><span class="mouse" />
      </template>
      <template v-else><span class="notebook" /><span class="pencil" /></template>
      <span v-if="power" class="power-strip"><i /><i /></span>
    </span>
    <span class="desk-chair" />
  </span>
</template>

<style scoped>
.plan-desk { position: relative; display: block; width: 66px; height: 52px; flex: 0 0 auto; filter: drop-shadow(0 3px 3px #0002); }
.desk-top { position: absolute; inset: 3px 3px 14px; border: 1px solid #b49773; border-radius: 5px; background: repeating-linear-gradient(3deg, transparent 0 4px, #79522c0d 5px, transparent 6px 9px), linear-gradient(115deg, #eddbba, #cfa97c); box-shadow: inset 0 1px 0 #fff9, inset 0 -3px 0 #986b4140, 0 2px 0 #8e7457; }
.desk-legs { position: absolute; inset: 7px 7px 11px; border-inline: 3px solid #8b98a2; border-radius: 2px; background: linear-gradient(90deg, #d2dce4, #71818b); }
.desk-chair { position: absolute; width: 27px; height: 16px; bottom: 1px; left: 20px; border-radius: 5px 5px 9px 9px; border: 1px solid #526d6a; background: linear-gradient(#779d94, #567c76); box-shadow: inset 0 -4px 0 #395a5655, inset 0 1px 0 #fff5, 0 2px 2px #0002; }
.notebook { position: absolute; width: 21px; height: 22px; left: 13px; top: 5px; border: 1px solid #c4ba9f; border-radius: 2px; transform: rotate(-8deg); background: linear-gradient(90deg, transparent 47%, #b9b19e 49% 51%, transparent 53%), repeating-linear-gradient(0deg, #f4efdf 0 4px, #ded7c5 4px 5px); box-shadow: 1px 2px 1px #61472526; }
.pencil { position: absolute; width: 2px; height: 19px; left: 40px; top: 9px; transform: rotate(13deg); background: linear-gradient(#5b6d7a 0 12%, #a67742 12% 85%, #444 85%); }
.power-strip { position: absolute; display: flex; gap: 3px; right: 3px; top: 2px; padding: 2px; border: 1px solid #acaaa3; border-radius: 2px; background: linear-gradient(#fff, #d9d8d1); box-shadow: 0 1px 2px #0002; }
.power-strip i { width: 3px; height: 4px; border-inline: 1px solid #58616a; }
.monitor { position: absolute; width: 29px; height: 18px; top: 1px; left: 10px; padding: 2px; border-radius: 3px; background: linear-gradient(#586773, #26333d); box-shadow: 0 2px 2px #0004; }
.monitor span { display: block; height: 100%; border-radius: 1px; background: linear-gradient(140deg, #d1f3f2 0%, #69b8c6 40%, #468494 42%, #8bd3d8); box-shadow: inset 0 0 2px #ffffff80; }
.monitor::after { content: ''; position: absolute; left: 12px; bottom: -3px; width: 6px; height: 3px; background: #71808c; }
.keyboard { position: absolute; width: 24px; height: 7px; bottom: 4px; left: 12px; border: 1px solid #9da7ab; border-radius: 2px; background: repeating-linear-gradient(90deg, #dce4e5 0 2px, #8e9ba3 2px 3px); }
.mouse { position: absolute; width: 5px; height: 8px; bottom: 4px; right: 13px; border: 1px solid #b0b8ba; border-radius: 4px; background: linear-gradient(90deg, #f1f4f3, #aab7bd); }
:global([data-theme='dark']) .desk-top { background: repeating-linear-gradient(3deg, transparent 0 4px, #311b121a 5px, transparent 6px 9px), linear-gradient(115deg, #b99b76, #886849); border-color: #ae8a63; box-shadow: inset 0 1px 0 #f4dab959, inset 0 -3px 0 #533a3150, 0 2px 0 #49392c; }
</style>
